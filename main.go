package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"net/http"
	"net/url"

	"docker.io/go-docker"
	"docker.io/go-docker/api/types"
	"docker.io/go-docker/api/types/registry"
	"github.com/c-bata/go-prompt"
	"github.com/patrickmn/go-cache"
)

var dockerClient *docker.Client
var lastValidKeyword string
var subCommands = map[string][]prompt.Suggest{
	"attach": {
		prompt.Suggest{Text: "--detach-keys", Description: "Override the key sequence for detaching a container"},
		prompt.Suggest{Text: "--no-stdin", Description: "Do not attach STDIN"},
		prompt.Suggest{Text: "--sig-proxy", Description: "Proxy all received signals to the process"},
	},
	"build": {
		prompt.Suggest{Text: "--add-host", Description: "Add a custom host-to-IP mapping (host:ip)"},
		prompt.Suggest{Text: "--build-arg", Description: "Set build-time variables"},
		prompt.Suggest{Text: "--cache-from", Description: "Images to consider as cache sources"},
		prompt.Suggest{Text: "--cgroup-parent", Description: "Optional parent cgroup for the container"},
		prompt.Suggest{Text: "--compress", Description: "Compress the build context using gzip"},
		prompt.Suggest{Text: "--cpu-period", Description: "Limit the CPU CFS (Completely Fair Scheduler) period"},
		prompt.Suggest{Text: "--cpu-quota", Description: "Limit the CPU CFS (Completely Fair Scheduler) quota"},
		prompt.Suggest{Text: "--cpu-shares", Description: "CPU shares (relative weight)"},
		prompt.Suggest{Text: "--cpuset-cpus", Description: "CPUs in which to allow execution (0-3, 0,1)"},
		prompt.Suggest{Text: "--cpuset-mems", Description: "MEMs in which to allow execution (0-3, 0,1)"},
		prompt.Suggest{Text: "--disable-content-trust", Description: "Skip image verification"},
		prompt.Suggest{Text: "--file", Description: "Name of the Dockerfile (Default is ‘PATH/Dockerfile’)"},
		prompt.Suggest{Text: "--force-rm", Description: "Always remove intermediate containers"},
		prompt.Suggest{Text: "--iidfile", Description: "Write the image ID to the file"},
		prompt.Suggest{Text: "--isolation", Description: "Container isolation technology"},
		prompt.Suggest{Text: "--label", Description: "Set metadata for an image"},
		prompt.Suggest{Text: "--memory", Description: "Memory limit"},
		prompt.Suggest{Text: "--memory-swap", Description: "Swap limit equal to memory plus swap: ‘-1’ to enable unlimited swap"},
		prompt.Suggest{Text: "--network", Description: ""},
		prompt.Suggest{Text: "--no-cache", Description: "Do not use cache when building the image"},
		prompt.Suggest{Text: "--output", Description: ""},
		prompt.Suggest{Text: "--platform", Description: ""},
		prompt.Suggest{Text: "--progress", Description: "Set type of progress output (auto, plain, tty). Use plain to show container output"},
		prompt.Suggest{Text: "--pull", Description: "Always attempt to pull a newer version of the image"},
		prompt.Suggest{Text: "--quiet", Description: "Suppress the build output and print image ID on success"},
		prompt.Suggest{Text: "--rm", Description: "Remove intermediate containers after a successful build"},
		prompt.Suggest{Text: "--secret", Description: ""},
		prompt.Suggest{Text: "--security-opt", Description: "Security options"},
		prompt.Suggest{Text: "--shm-size", Description: "Size of /dev/shm"},
		prompt.Suggest{Text: "--squash", Description: ""},
		prompt.Suggest{Text: "--ssh", Description: ""},
		prompt.Suggest{Text: "--stream", Description: ""},
		prompt.Suggest{Text: "--tag", Description: "Name and optionally a tag in the ‘name:tag’ format"},
		prompt.Suggest{Text: "--target", Description: "Set the target build stage to build."},
		prompt.Suggest{Text: "--ulimit", Description: "Ulimit options"},
	},
	"commit": {
		prompt.Suggest{Text: "--author", Description: "Author (e.g., “John Hannibal Smith "},
		prompt.Suggest{Text: "--change", Description: "Apply Dockerfile instruction to the created image"},
		prompt.Suggest{Text: "--message", Description: "Commit message"},
		prompt.Suggest{Text: "--pause", Description: "Pause container during commit"},
	},
	"cp": {
		prompt.Suggest{Text: "--archive", Description: "Archive mode (copy all uid/gid information)"},
		prompt.Suggest{Text: "--follow-link", Description: "Always follow symbol link in SRC_PATH"},
	},
	"create": {
		prompt.Suggest{Text: "--add-host", Description: "Add a custom host-to-IP mapping (host:ip)"},
		prompt.Suggest{Text: "--attach", Description: "Attach to STDIN, STDOUT or STDERR"},
		prompt.Suggest{Text: "--blkio-weight", Description: "Block IO (relative weight), between 10 and 1000, or 0 to disable (default 0)"},
		prompt.Suggest{Text: "--blkio-weight-device", Description: "Block IO weight (relative device weight)"},
		prompt.Suggest{Text: "--cap-add", Description: "Add Linux capabilities"},
		prompt.Suggest{Text: "--cap-drop", Description: "Drop Linux capabilities"},
		prompt.Suggest{Text: "--cgroup-parent", Description: "Optional parent cgroup for the container"},
		prompt.Suggest{Text: "--cidfile", Description: "Write the container ID to the file"},
		prompt.Suggest{Text: "--cpu-count", Description: "CPU count (Windows only)"},
		prompt.Suggest{Text: "--cpu-percent", Description: "CPU percent (Windows only)"},
		prompt.Suggest{Text: "--cpu-period", Description: "Limit CPU CFS (Completely Fair Scheduler) period"},
		prompt.Suggest{Text: "--cpu-quota", Description: "Limit CPU CFS (Completely Fair Scheduler) quota"},
		prompt.Suggest{Text: "--cpu-rt-period", Description: ""},
		prompt.Suggest{Text: "--cpu-rt-runtime", Description: ""},
		prompt.Suggest{Text: "--cpu-shares", Description: "CPU shares (relative weight)"},
		prompt.Suggest{Text: "--cpus", Description: ""},
		prompt.Suggest{Text: "--cpuset-cpus", Description: "CPUs in which to allow execution (0-3, 0,1)"},
		prompt.Suggest{Text: "--cpuset-mems", Description: "MEMs in which to allow execution (0-3, 0,1)"},
		prompt.Suggest{Text: "--device", Description: "Add a host device to the container"},
		prompt.Suggest{Text: "--device-cgroup-rule", Description: "Add a rule to the cgroup allowed devices list"},
		prompt.Suggest{Text: "--device-read-bps", Description: "Limit read rate (bytes per second) from a device"},
		prompt.Suggest{Text: "--device-read-iops", Description: "Limit read rate (IO per second) from a device"},
		prompt.Suggest{Text: "--device-write-bps", Description: "Limit write rate (bytes per second) to a device"},
		prompt.Suggest{Text: "--device-write-iops", Description: "Limit write rate (IO per second) to a device"},
		prompt.Suggest{Text: "--disable-content-trust", Description: "Skip image verification"},
		prompt.Suggest{Text: "--dns", Description: "Set custom DNS servers"},
		prompt.Suggest{Text: "--dns-opt", Description: "Set DNS options"},
		prompt.Suggest{Text: "--dns-option", Description: "Set DNS options"},
		prompt.Suggest{Text: "--dns-search", Description: "Set custom DNS search domains"},
		prompt.Suggest{Text: "--domainname", Description: "Container NIS domain name"},
		prompt.Suggest{Text: "--entrypoint", Description: "Overwrite the default ENTRYPOINT of the image"},
		prompt.Suggest{Text: "--env", Description: "Set environment variables"},
		prompt.Suggest{Text: "--env-file", Description: "Read in a file of environment variables"},
		prompt.Suggest{Text: "--expose", Description: "Expose a port or a range of ports"},
		prompt.Suggest{Text: "--gpus", Description: ""},
		prompt.Suggest{Text: "--group-add", Description: "Add additional groups to join"},
		prompt.Suggest{Text: "--health-cmd", Description: "Command to run to check health"},
		prompt.Suggest{Text: "--health-interval", Description: "Time between running the check (ms|s|m|h) (default 0s)"},
		prompt.Suggest{Text: "--health-retries", Description: "Consecutive failures needed to report unhealthy"},
		prompt.Suggest{Text: "--health-start-period", Description: ""},
		prompt.Suggest{Text: "--health-timeout", Description: "Maximum time to allow one check to run (ms|s|m|h) (default 0s)"},
		prompt.Suggest{Text: "--help", Description: "Print usage"},
		prompt.Suggest{Text: "--hostname", Description: "Container host name"},
		prompt.Suggest{Text: "--init", Description: ""},
		prompt.Suggest{Text: "--interactive", Description: "Keep STDIN open even if not attached"},
		prompt.Suggest{Text: "--io-maxbandwidth", Description: "Maximum IO bandwidth limit for the system drive (Windows only)"},
		prompt.Suggest{Text: "--io-maxiops", Description: "Maximum IOps limit for the system drive (Windows only)"},
		prompt.Suggest{Text: "--ip", Description: "IPv4 address (e.g., 172.30.100.104)"},
		prompt.Suggest{Text: "--ip6", Description: "IPv6 address (e.g., 2001:db8::33)"},
		prompt.Suggest{Text: "--ipc", Description: "IPC mode to use"},
		prompt.Suggest{Text: "--isolation", Description: "Container isolation technology"},
		prompt.Suggest{Text: "--kernel-memory", Description: "Kernel memory limit"},
		prompt.Suggest{Text: "--label", Description: "Set meta data on a container"},
		prompt.Suggest{Text: "--label-file", Description: "Read in a line delimited file of labels"},
		prompt.Suggest{Text: "--link", Description: "Add link to another container"},
		prompt.Suggest{Text: "--link-local-ip", Description: "Container IPv4/IPv6 link-local addresses"},
		prompt.Suggest{Text: "--log-driver", Description: "Logging driver for the container"},
		prompt.Suggest{Text: "--log-opt", Description: "Log driver options"},
		prompt.Suggest{Text: "--mac-address", Description: "Container MAC address (e.g., 92:d0:c6:0a:29:33)"},
		prompt.Suggest{Text: "--memory", Description: "Memory limit"},
		prompt.Suggest{Text: "--memory-reservation", Description: "Memory soft limit"},
		prompt.Suggest{Text: "--memory-swap", Description: "Swap limit equal to memory plus swap: ‘-1’ to enable unlimited swap"},
		prompt.Suggest{Text: "--memory-swappiness", Description: "Tune container memory swappiness (0 to 100)"},
		prompt.Suggest{Text: "--mount", Description: "Attach a filesystem mount to the container"},
		prompt.Suggest{Text: "--name", Description: "Assign a name to the container"},
		prompt.Suggest{Text: "--net", Description: "Connect a container to a network"},
		prompt.Suggest{Text: "--net-alias", Description: "Add network-scoped alias for the container"},
		prompt.Suggest{Text: "--network", Description: "Connect a container to a network"},
		prompt.Suggest{Text: "--network-alias", Description: "Add network-scoped alias for the container"},
		prompt.Suggest{Text: "--no-healthcheck", Description: "Disable any container-specified HEALTHCHECK"},
		prompt.Suggest{Text: "--oom-kill-disable", Description: "Disable OOM Killer"},
		prompt.Suggest{Text: "--oom-score-adj", Description: "Tune host’s OOM preferences (-1000 to 1000)"},
		prompt.Suggest{Text: "--pid", Description: "PID namespace to use"},
		prompt.Suggest{Text: "--pids-limit", Description: "Tune container pids limit (set -1 for unlimited)"},
		prompt.Suggest{Text: "--platform", Description: ""},
		prompt.Suggest{Text: "--privileged", Description: "Give extended privileges to this container"},
		prompt.Suggest{Text: "--publish", Description: "Publish a container’s port(s) to the host"},
		prompt.Suggest{Text: "--publish-all", Description: "Publish all exposed ports to random ports"},
		prompt.Suggest{Text: "--read-only", Description: "Mount the container’s root filesystem as read only"},
		prompt.Suggest{Text: "--restart", Description: "Restart policy to apply when a container exits"},
		prompt.Suggest{Text: "--rm", Description: "Automatically remove the container when it exits"},
		prompt.Suggest{Text: "--runtime", Description: "Runtime to use for this container"},
		prompt.Suggest{Text: "--security-opt", Description: "Security Options"},
		prompt.Suggest{Text: "--shm-size", Description: "Size of /dev/shm"},
		prompt.Suggest{Text: "--stop-signal", Description: "Signal to stop a container"},
		prompt.Suggest{Text: "--stop-timeout", Description: ""},
		prompt.Suggest{Text: "--storage-opt", Description: "Storage driver options for the container"},
		prompt.Suggest{Text: "--sysctl", Description: "Sysctl options"},
		prompt.Suggest{Text: "--tmpfs", Description: "Mount a tmpfs directory"},
		prompt.Suggest{Text: "--tty", Description: "Allocate a pseudo-TTY"},
		prompt.Suggest{Text: "--ulimit", Description: "Ulimit options"},
		prompt.Suggest{Text: "--user", Description: "Username or UID (format: &lt;name|uid&gt;[:&lt;group|gid&gt;])"},
		prompt.Suggest{Text: "--userns", Description: "User namespace to use"},
		prompt.Suggest{Text: "--uts", Description: "UTS namespace to use"},
		prompt.Suggest{Text: "--volume", Description: "Bind mount a volume"},
		prompt.Suggest{Text: "--volume-driver", Description: "Optional volume driver for the container"},
		prompt.Suggest{Text: "--volumes-from", Description: "Mount volumes from the specified container(s)"},
		prompt.Suggest{Text: "--workdir", Description: "Working directory inside the container"},
	},
	"events": {
		prompt.Suggest{Text: "--filter", Description: "Filter output based on conditions provided"},
		prompt.Suggest{Text: "--format", Description: "Format the output using the given Go template"},
		prompt.Suggest{Text: "--since", Description: "Show all events created since timestamp"},
		prompt.Suggest{Text: "--until", Description: "Stream events until this timestamp"},
	},
	"exec": {
		prompt.Suggest{Text: "--detach", Description: "Detached mode: run command in the background"},
		prompt.Suggest{Text: "--detach-keys", Description: "Override the key sequence for detaching a container"},
		prompt.Suggest{Text: "--env", Description: ""},
		prompt.Suggest{Text: "--interactive", Description: "Keep STDIN open even if not attached"},
		prompt.Suggest{Text: "--privileged", Description: "Give extended privileges to the command"},
		prompt.Suggest{Text: "--tty", Description: "Allocate a pseudo-TTY"},
		prompt.Suggest{Text: "--user", Description: "Username or UID (format: &lt;name|uid&gt;[:&lt;group|gid&gt;])"},
		prompt.Suggest{Text: "--workdir", Description: ""},
	},
	"export": {
		prompt.Suggest{Text: "--output", Description: "Write to a file, instead of STDOUT"},
	},
	"history": {
		prompt.Suggest{Text: "--format", Description: "Pretty-print images using a Go template"},
		prompt.Suggest{Text: "--human", Description: "Print sizes and dates in human readable format"},
		prompt.Suggest{Text: "--no-trunc", Description: "Don’t truncate output"},
		prompt.Suggest{Text: "--quiet", Description: "Only show numeric IDs"},
	},
	"images": {
		prompt.Suggest{Text: "--all", Description: "Show all images (default hides intermediate images)"},
		prompt.Suggest{Text: "--digests", Description: "Show digests"},
		prompt.Suggest{Text: "--filter", Description: "Filter output based on conditions provided"},
		prompt.Suggest{Text: "--format", Description: "Pretty-print images using a Go template"},
		prompt.Suggest{Text: "--no-trunc", Description: "Don’t truncate output"},
		prompt.Suggest{Text: "--quiet", Description: "Only show numeric IDs"},
	},
	"import": {
		prompt.Suggest{Text: "--change", Description: "Apply Dockerfile instruction to the created image"},
		prompt.Suggest{Text: "--message", Description: "Set commit message for imported image"},
		prompt.Suggest{Text: "--platform", Description: ""},
	},
	"info": {
		prompt.Suggest{Text: "--format", Description: "Format the output using the given Go template"},
	},
	"inspect": {
		prompt.Suggest{Text: "--format", Description: "Format the output using the given Go template"},
		prompt.Suggest{Text: "--size", Description: "Display total file sizes if the type is container"},
		prompt.Suggest{Text: "--type", Description: "Return JSON for specified type"},
	},
	"kill": {
		prompt.Suggest{Text: "--signal", Description: "Signal to send to the container"},
	},
	"load": {
		prompt.Suggest{Text: "--input", Description: "Read from tar archive file, instead of STDIN"},
		prompt.Suggest{Text: "--quiet", Description: "Suppress the load output"},
	},
	"login": {
		prompt.Suggest{Text: "--password", Description: "Password"},
		prompt.Suggest{Text: "--password-stdin", Description: "Take the password from stdin"},
		prompt.Suggest{Text: "--username", Description: "Username"},
	},
	"logs": {
		prompt.Suggest{Text: "--details", Description: "Show extra details provided to logs"},
		prompt.Suggest{Text: "--follow", Description: "Follow log output"},
		prompt.Suggest{Text: "--since", Description: "Show logs since timestamp (e.g. 2013-01-02T13:23:37) or relative (e.g. 42m for 42 minutes)"},
		prompt.Suggest{Text: "--tail", Description: "Number of lines to show from the end of the logs"},
		prompt.Suggest{Text: "--timestamps", Description: "Show timestamps"},
		prompt.Suggest{Text: "--until", Description: ""},
	},
	"ps": {
		prompt.Suggest{Text: "--all", Description: "Show all containers (default shows just running)"},
		prompt.Suggest{Text: "--filter", Description: "Filter output based on conditions provided"},
		prompt.Suggest{Text: "--format", Description: "Pretty-print containers using a Go template"},
		prompt.Suggest{Text: "--last", Description: "Show n last created containers (includes all states)"},
		prompt.Suggest{Text: "--latest", Description: "Show the latest created container (includes all states)"},
		prompt.Suggest{Text: "--no-trunc", Description: "Don’t truncate output"},
		prompt.Suggest{Text: "--quiet", Description: "Only display numeric IDs"},
		prompt.Suggest{Text: "--size", Description: "Display total file sizes"},
	},
	"pull": {
		prompt.Suggest{Text: "--all-tags", Description: "Download all tagged images in the repository"},
		prompt.Suggest{Text: "--disable-content-trust", Description: "Skip image verification"},
		prompt.Suggest{Text: "--platform", Description: ""},
		prompt.Suggest{Text: "--quiet", Description: "Suppress verbose output"},
	},
	"push": {
		prompt.Suggest{Text: "--disable-content-trust", Description: "Skip image signing"},
	},
	"restart": {
		prompt.Suggest{Text: "--time", Description: "Seconds to wait for stop before killing the container"},
	},
	"rm": {
		prompt.Suggest{Text: "--force", Description: "Force the removal of a running container (uses SIGKILL)"},
		prompt.Suggest{Text: "--link", Description: "Remove the specified link"},
		prompt.Suggest{Text: "--volumes", Description: "Remove the volumes associated with the container"},
	},
	"rmi": {
		prompt.Suggest{Text: "--force", Description: "Force removal of the image"},
		prompt.Suggest{Text: "--no-prune", Description: "Do not delete untagged parents"},
	},
	"run": {
		prompt.Suggest{Text: "--add-host", Description: "Add a custom host-to-IP mapping (host:ip)"},
		prompt.Suggest{Text: "--attach", Description: "Attach to STDIN, STDOUT or STDERR"},
		prompt.Suggest{Text: "--blkio-weight", Description: "Block IO (relative weight), between 10 and 1000, or 0 to disable (default 0)"},
		prompt.Suggest{Text: "--blkio-weight-device", Description: "Block IO weight (relative device weight)"},
		prompt.Suggest{Text: "--cap-add", Description: "Add Linux capabilities"},
		prompt.Suggest{Text: "--cap-drop", Description: "Drop Linux capabilities"},
		prompt.Suggest{Text: "--cgroup-parent", Description: "Optional parent cgroup for the container"},
		prompt.Suggest{Text: "--cidfile", Description: "Write the container ID to the file"},
		prompt.Suggest{Text: "--cpu-count", Description: "CPU count (Windows only)"},
		prompt.Suggest{Text: "--cpu-percent", Description: "CPU percent (Windows only)"},
		prompt.Suggest{Text: "--cpu-period", Description: "Limit CPU CFS (Completely Fair Scheduler) period"},
		prompt.Suggest{Text: "--cpu-quota", Description: "Limit CPU CFS (Completely Fair Scheduler) quota"},
		prompt.Suggest{Text: "--cpu-rt-period", Description: ""},
		prompt.Suggest{Text: "--cpu-rt-runtime", Description: ""},
		prompt.Suggest{Text: "--cpu-shares", Description: "CPU shares (relative weight)"},
		prompt.Suggest{Text: "--cpus", Description: ""},
		prompt.Suggest{Text: "--cpuset-cpus", Description: "CPUs in which to allow execution (0-3, 0,1)"},
		prompt.Suggest{Text: "--cpuset-mems", Description: "MEMs in which to allow execution (0-3, 0,1)"},
		prompt.Suggest{Text: "--detach", Description: "Run container in background and print container ID"},
		prompt.Suggest{Text: "--detach-keys", Description: "Override the key sequence for detaching a container"},
		prompt.Suggest{Text: "--device", Description: "Add a host device to the container"},
		prompt.Suggest{Text: "--device-cgroup-rule", Description: "Add a rule to the cgroup allowed devices list"},
		prompt.Suggest{Text: "--device-read-bps", Description: "Limit read rate (bytes per second) from a device"},
		prompt.Suggest{Text: "--device-read-iops", Description: "Limit read rate (IO per second) from a device"},
		prompt.Suggest{Text: "--device-write-bps", Description: "Limit write rate (bytes per second) to a device"},
		prompt.Suggest{Text: "--device-write-iops", Description: "Limit write rate (IO per second) to a device"},
		prompt.Suggest{Text: "--disable-content-trust", Description: "Skip image verification"},
		prompt.Suggest{Text: "--dns", Description: "Set custom DNS servers"},
		prompt.Suggest{Text: "--dns-opt", Description: "Set DNS options"},
		prompt.Suggest{Text: "--dns-option", Description: "Set DNS options"},
		prompt.Suggest{Text: "--dns-search", Description: "Set custom DNS search domains"},
		prompt.Suggest{Text: "--domainname", Description: "Container NIS domain name"},
		prompt.Suggest{Text: "--entrypoint", Description: "Overwrite the default ENTRYPOINT of the image"},
		prompt.Suggest{Text: "--env", Description: "Set environment variables"},
		prompt.Suggest{Text: "--env-file", Description: "Read in a file of environment variables"},
		prompt.Suggest{Text: "--expose", Description: "Expose a port or a range of ports"},
		prompt.Suggest{Text: "--gpus", Description: ""},
		prompt.Suggest{Text: "--group-add", Description: "Add additional groups to join"},
		prompt.Suggest{Text: "--health-cmd", Description: "Command to run to check health"},
		prompt.Suggest{Text: "--health-interval", Description: "Time between running the check (ms|s|m|h) (default 0s)"},
		prompt.Suggest{Text: "--health-retries", Description: "Consecutive failures needed to report unhealthy"},
		prompt.Suggest{Text: "--health-start-period", Description: ""},
		prompt.Suggest{Text: "--health-timeout", Description: "Maximum time to allow one check to run (ms|s|m|h) (default 0s)"},
		prompt.Suggest{Text: "--help", Description: "Print usage"},
		prompt.Suggest{Text: "--hostname", Description: "Container host name"},
		prompt.Suggest{Text: "--init", Description: ""},
		prompt.Suggest{Text: "--interactive", Description: "Keep STDIN open even if not attached"},
		prompt.Suggest{Text: "--io-maxbandwidth", Description: "Maximum IO bandwidth limit for the system drive (Windows only)"},
		prompt.Suggest{Text: "--io-maxiops", Description: "Maximum IOps limit for the system drive (Windows only)"},
		prompt.Suggest{Text: "--ip", Description: "IPv4 address (e.g., 172.30.100.104)"},
		prompt.Suggest{Text: "--ip6", Description: "IPv6 address (e.g., 2001:db8::33)"},
		prompt.Suggest{Text: "--ipc", Description: "IPC mode to use"},
		prompt.Suggest{Text: "--isolation", Description: "Container isolation technology"},
		prompt.Suggest{Text: "--kernel-memory", Description: "Kernel memory limit"},
		prompt.Suggest{Text: "--label", Description: "Set meta data on a container"},
		prompt.Suggest{Text: "--label-file", Description: "Read in a line delimited file of labels"},
		prompt.Suggest{Text: "--link", Description: "Add link to another container"},
		prompt.Suggest{Text: "--link-local-ip", Description: "Container IPv4/IPv6 link-local addresses"},
		prompt.Suggest{Text: "--log-driver", Description: "Logging driver for the container"},
		prompt.Suggest{Text: "--log-opt", Description: "Log driver options"},
		prompt.Suggest{Text: "--mac-address", Description: "Container MAC address (e.g., 92:d0:c6:0a:29:33)"},
		prompt.Suggest{Text: "--memory", Description: "Memory limit"},
		prompt.Suggest{Text: "--memory-reservation", Description: "Memory soft limit"},
		prompt.Suggest{Text: "--memory-swap", Description: "Swap limit equal to memory plus swap: ‘-1’ to enable unlimited swap"},
		prompt.Suggest{Text: "--memory-swappiness", Description: "Tune container memory swappiness (0 to 100)"},
		prompt.Suggest{Text: "--mount", Description: "Attach a filesystem mount to the container"},
		prompt.Suggest{Text: "--name", Description: "Assign a name to the container"},
		prompt.Suggest{Text: "--net", Description: "Connect a container to a network"},
		prompt.Suggest{Text: "--net-alias", Description: "Add network-scoped alias for the container"},
		prompt.Suggest{Text: "--network", Description: "Connect a container to a network"},
		prompt.Suggest{Text: "--network-alias", Description: "Add network-scoped alias for the container"},
		prompt.Suggest{Text: "--no-healthcheck", Description: "Disable any container-specified HEALTHCHECK"},
		prompt.Suggest{Text: "--oom-kill-disable", Description: "Disable OOM Killer"},
		prompt.Suggest{Text: "--oom-score-adj", Description: "Tune host’s OOM preferences (-1000 to 1000)"},
		prompt.Suggest{Text: "--pid", Description: "PID namespace to use"},
		prompt.Suggest{Text: "--pids-limit", Description: "Tune container pids limit (set -1 for unlimited)"},
		prompt.Suggest{Text: "--platform", Description: ""},
		prompt.Suggest{Text: "--privileged", Description: "Give extended privileges to this container"},
		prompt.Suggest{Text: "--publish", Description: "Publish a container’s port(s) to the host"},
		prompt.Suggest{Text: "--publish-all", Description: "Publish all exposed ports to random ports"},
		prompt.Suggest{Text: "--read-only", Description: "Mount the container’s root filesystem as read only"},
		prompt.Suggest{Text: "--restart", Description: "Restart policy to apply when a container exits"},
		prompt.Suggest{Text: "--rm", Description: "Automatically remove the container when it exits"},
		prompt.Suggest{Text: "--runtime", Description: "Runtime to use for this container"},
		prompt.Suggest{Text: "--security-opt", Description: "Security Options"},
		prompt.Suggest{Text: "--shm-size", Description: "Size of /dev/shm"},
		prompt.Suggest{Text: "--sig-proxy", Description: "Proxy received signals to the process"},
		prompt.Suggest{Text: "--stop-signal", Description: "Signal to stop a container"},
		prompt.Suggest{Text: "--stop-timeout", Description: ""},
		prompt.Suggest{Text: "--storage-opt", Description: "Storage driver options for the container"},
		prompt.Suggest{Text: "--sysctl", Description: "Sysctl options"},
		prompt.Suggest{Text: "--tmpfs", Description: "Mount a tmpfs directory"},
		prompt.Suggest{Text: "--tty", Description: "Allocate a pseudo-TTY"},
		prompt.Suggest{Text: "--ulimit", Description: "Ulimit options"},
		prompt.Suggest{Text: "--user", Description: "Username or UID (format: &lt;name|uid&gt;[:&lt;group|gid&gt;])"},
		prompt.Suggest{Text: "--userns", Description: "User namespace to use"},
		prompt.Suggest{Text: "--uts", Description: "UTS namespace to use"},
		prompt.Suggest{Text: "--volume", Description: "Bind mount a volume"},
		prompt.Suggest{Text: "--volume-driver", Description: "Optional volume driver for the container"},
		prompt.Suggest{Text: "--volumes-from", Description: "Mount volumes from the specified container(s)"},
		prompt.Suggest{Text: "--workdir", Description: "Working directory inside the container"},
	},
	"save": {
		prompt.Suggest{Text: "--output", Description: "Write to a file, instead of STDOUT"},
	},
	"search": {
		prompt.Suggest{Text: "--automated", Description: ""},
		prompt.Suggest{Text: "--filter", Description: "Filter output based on conditions provided"},
		prompt.Suggest{Text: "--format", Description: "Pretty-print search using a Go template"},
		prompt.Suggest{Text: "--limit", Description: "Max number of search results"},
		prompt.Suggest{Text: "--no-trunc", Description: "Don’t truncate output"},
		prompt.Suggest{Text: "--stars", Description: ""},
	},
	"stack": {
		prompt.Suggest{Text: "--kubeconfig", Description: ""},
		prompt.Suggest{Text: "--orchestrator", Description: "Orchestrator to use (swarm|kubernetes|all)"},
	},
	"start": {
		prompt.Suggest{Text: "--attach", Description: "Attach STDOUT/STDERR and forward signals"},
		prompt.Suggest{Text: "--checkpoint", Description: ""},
		prompt.Suggest{Text: "--checkpoint-dir", Description: ""},
		prompt.Suggest{Text: "--detach-keys", Description: "Override the key sequence for detaching a container"},
		prompt.Suggest{Text: "--interactive", Description: "Attach container’s STDIN"},
	},
	"stats": {
		prompt.Suggest{Text: "--all", Description: "Show all containers (default shows just running)"},
		prompt.Suggest{Text: "--format", Description: "Pretty-print images using a Go template"},
		prompt.Suggest{Text: "--no-stream", Description: "Disable streaming stats and only pull the first result"},
		prompt.Suggest{Text: "--no-trunc", Description: "Do not truncate output"},
	},
	"stop": {
		prompt.Suggest{Text: "--time", Description: "Seconds to wait for stop before killing it"},
	},
	"update": {
		prompt.Suggest{Text: "--blkio-weight", Description: "Block IO (relative weight), between 10 and 1000, or 0 to disable (default 0)"},
		prompt.Suggest{Text: "--cpu-period", Description: "Limit CPU CFS (Completely Fair Scheduler) period"},
		prompt.Suggest{Text: "--cpu-quota", Description: "Limit CPU CFS (Completely Fair Scheduler) quota"},
		prompt.Suggest{Text: "--cpu-rt-period", Description: ""},
		prompt.Suggest{Text: "--cpu-rt-runtime", Description: ""},
		prompt.Suggest{Text: "--cpu-shares", Description: "CPU shares (relative weight)"},
		prompt.Suggest{Text: "--cpus", Description: ""},
		prompt.Suggest{Text: "--cpuset-cpus", Description: "CPUs in which to allow execution (0-3, 0,1)"},
		prompt.Suggest{Text: "--cpuset-mems", Description: "MEMs in which to allow execution (0-3, 0,1)"},
		prompt.Suggest{Text: "--kernel-memory", Description: "Kernel memory limit"},
		prompt.Suggest{Text: "--memory", Description: "Memory limit"},
		prompt.Suggest{Text: "--memory-reservation", Description: "Memory soft limit"},
		prompt.Suggest{Text: "--memory-swap", Description: "Swap limit equal to memory plus swap: ‘-1’ to enable unlimited swap"},
		prompt.Suggest{Text: "--pids-limit", Description: ""},
		prompt.Suggest{Text: "--restart", Description: "Restart policy to apply when a container exits"},
	},
	"version": {
		prompt.Suggest{Text: "--format", Description: "Format the output using the given Go template"},
		prompt.Suggest{Text: "--kubeconfig", Description: ""},
	},
}

func isDockerCommand(kw string) bool {
	dockerCommands := []string{
		"docker",
		"attach",
		"build",
		"builder",
		"checkpoint",
		"commit",
		"config",
		"container",
		"context",
		"cp",
		"create",
		"diff",
		"events",
		"exec",
		"export",
		"history",
		"image",
		"images",
		"import",
		"info",
		"inspect",
		"kill",
		"load",
		"login",
		"logout",
		"logs",
		"manifest",
		"network",
		"node",
		"pause",
		"plugin",
		"port",
		"ps",
		"pull",
		"push",
		"rename",
		"restart",
		"rm",
		"rmi",
		"run",
		"save",
		"search",
		"secret",
		"service",
		"stack",
		"start",
		"stats",
		"stop",
		"swarm",
		"system",
		"tag",
		"top",
		"trust",
		"unpause",
		"update",
		"version",
		"volume",
		"wait",
	}

	for _, cmd := range dockerCommands {
		if cmd == kw {
			return true
		}
	}

	return false
}

//DockerHubResult : Wrap DockerHub API call
type DockerHubResult struct {
	PageCount        *int                    `json:"num_pages,omitempty"`
	ResultCount      *int                    `json:"num_results,omitempty"`
	ItemCountPerPage *int                    `json:"page_size,omitempty"`
	CurrentPage      *int                    `json:"page,omitempty"`
	Query            *string                 `json:"query,omitempty"`
	Items            []registry.SearchResult `json:"results,omitempty"`
}

func imageFromHubAPI(count int) []registry.SearchResult {
	url := url.URL{
		Scheme:   "https",
		Host:     "registry.hub.docker.com",
		Path:     "/v2/repositories/library",
		RawQuery: "page=1&page_size=" + strconv.Itoa(count),
	}

	client := &http.Client{
		Timeout: 2 * time.Second,
	}
	apiURL := url.String()

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil
	}
	req.Header.Add("Content-Type", "application/json")
	response, err := client.Do(req)
	if err != nil {
		return nil
	}

	defer response.Body.Close()

	decoder := json.NewDecoder(response.Body)
	searchResult := &DockerHubResult{}
	decoder.Decode(searchResult)
	return searchResult.Items
}

func imageFromContext(imageName string, count int) []registry.SearchResult {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	ctxResponse, err := dockerClient.ImageSearch(ctx, imageName, types.ImageSearchOptions{Limit: count})
	if err != nil {
		return nil
	}
	return ctxResponse
}

func imageFetchCompleter(imageName string, count int) []prompt.Suggest {
	searchResult := []registry.SearchResult{}
	if imageName == "" {
		searchResult = imageFromHubAPI(10)
	} else {
		searchResult = imageFromContext(imageName, 10)
	}

	suggestions := []prompt.Suggest{}
	for _, s := range searchResult {
		description := "Not Official"
		if s.IsOfficial {
			description = "Official"
		}
		suggestions = append(suggestions, prompt.Suggest{Text: s.Name, Description: "(" + description + ") " + s.Description})
	}
	return suggestions
}

var commandExpression = regexp.MustCompile(`(?P<command>exec|stop|start|service|pull|attach|build|commit|cp|create|events|export|history|images|import|info|inspect|kill|load|login|logs|ps|push|restart|rm|rmi|run|save|search|stack|stats|update|version)\s{1}`)

func getRegexGroups(text string) map[string]string {
	if !commandExpression.Match([]byte(text)) {
		return nil
	}

	match := commandExpression.FindStringSubmatch(text)
	result := make(map[string]string)
	for i, name := range commandExpression.SubexpNames() {
		if i != 0 && name != "" {
			result[name] = match[i]
		}
	}
	return result
}

var memoryCache = cache.New(5*time.Minute, 10*time.Minute)

func getFromCache(word string) []prompt.Suggest {
	cacheKey := "all"
	if word != "" {
		cacheKey = fmt.Sprintf("completer:%s", word)
	}
	completer, found := memoryCache.Get(cacheKey)
	if !found {
		completer = imageFetchCompleter(word, 10)
		memoryCache.Set(cacheKey, completer, cache.DefaultExpiration)
	}
	return completer.([]prompt.Suggest)
}

func completer(d prompt.Document) []prompt.Suggest {
	word := d.GetWordBeforeCursor()

	group := getRegexGroups(d.Text)
	if group != nil {
		command := group["command"]
		if command == "exec" || command == "stop" || command == "port" {
			return containerListCompleter(false)
		}

		if command == "start" {
			return containerListCompleter(true)
		}

		if command == "service" {
			return dockerServiceCommandCompleter()
		}

		if command == "pull" {
			if word != command {
				return getFromCache(word)
			}
			return getFromCache("")
		}
		if val, ok := subCommands[command]; ok {
			return prompt.FilterHasPrefix(val, word, true)
		}
	}

	suggestions := []prompt.Suggest{
		{Text: "attach", Description: "Attach local standard input, output, and error streams to a running container"},
		{Text: "build", Description: "Build an image from a Dockerfile"},
		{Text: "builder", Description: "Manage builds"},
		{Text: "checkpoint", Description: "Manage checkpoints"},
		{Text: "commit", Description: "Create a new image from a container’s changes"},
		{Text: "config", Description: "Manage Docker configs"},
		{Text: "container", Description: "Manage containers"},
		{Text: "context", Description: "Manage contexts"},
		{Text: "cp", Description: "Copy files/folders between a container and the local filesystem"},
		{Text: "create", Description: "Create a new container"},
		{Text: "diff", Description: "Inspect changes to files or directories on a container’s filesystem"},
		{Text: "events", Description: "Get real time events from the server"},
		{Text: "exec", Description: "Run a command in a running container"},
		{Text: "export", Description: "Export a container’s filesystem as a tar archive"},
		{Text: "history", Description: "Show the history of an image"},
		{Text: "image", Description: "Manage images"},
		{Text: "images", Description: "List images"},
		{Text: "import", Description: "Import the contents from a tarball to create a filesystem image"},
		{Text: "info", Description: "Display system-wide information"},
		{Text: "inspect", Description: "Return low-level information on Docker objects"},
		{Text: "kill", Description: "Kill one or more running containers"},
		{Text: "load", Description: "Load an image from a tar archive or STDIN"},
		{Text: "login", Description: "Log in to a Docker registry"},
		{Text: "logout", Description: "Log out from a Docker registry"},
		{Text: "logs", Description: "Fetch the logs of a container"},
		{Text: "manifest", Description: "Manage Docker image manifests and manifest lists"},
		{Text: "network", Description: "Manage networks"},
		{Text: "node", Description: "Manage Swarm nodes"},
		{Text: "pause", Description: "Pause all processes within one or more containers"},
		{Text: "plugin", Description: "Manage plugins"},
		{Text: "port", Description: "List port mappings or a specific mapping for the container"},
		{Text: "ps", Description: "List containers"},
		{Text: "pull", Description: "Pull an image or a repository from a registry"},
		{Text: "push", Description: "Push an image or a repository to a registry"},
		{Text: "rename", Description: "Rename a container"},
		{Text: "restart", Description: "Restart one or more containers"},
		{Text: "rm", Description: "Remove one or more containers"},
		{Text: "rmi", Description: "Remove one or more images"},
		{Text: "run", Description: "Run a command in a new container"},
		{Text: "save", Description: "Save one or more images to a tar archive (streamed to STDOUT by default)"},
		{Text: "search", Description: "Search the Docker Hub for images"},
		{Text: "secret", Description: "Manage Docker secrets"},
		{Text: "service", Description: "Manage services"},
		{Text: "stack", Description: "Manage Docker stacks"},
		{Text: "start", Description: "Start one or more stopped containers"},
		{Text: "stats", Description: "Display a live stream of container(s) resource usage statistics"},
		{Text: "stop", Description: "Stop one or more running containers"},
		{Text: "swarm", Description: "Manage Swarm"},
		{Text: "system", Description: "Manage Docker"},
		{Text: "tag", Description: "Create a tag TARGET_IMAGE that refers to SOURCE_IMAGE"},
		{Text: "top", Description: "Display the running processes of a container"},
		{Text: "trust", Description: "Manage trust on Docker images"},
		{Text: "unpause", Description: "Unpause all processes within one or more containers"},
		{Text: "update", Description: "Update configuration of one or more containers"},
		{Text: "version", Description: "Show the Docker version information"},
		{Text: "volume", Description: "Manage volumes"},
		{Text: "wait", Description: "Block until one or more containers stop, then print their exit codes"},
		{Text: "exit", Description: "Exit command prompt"},
	}

	return prompt.FilterHasPrefix(suggestions, word, true)
}

func dockerServiceCommandCompleter() []prompt.Suggest {
	return []prompt.Suggest{
		{Text: "create", Description: "Create a new service"},
		{Text: "inspect", Description: "Display detailed information on one or more services"},
		{Text: "logs", Description: "Fetch the logs of a service or task"},
		{Text: "ls", Description: "List services"},
		{Text: "ps", Description: "List the tasks of one or more services"},
		{Text: "rm", Description: "Remove one or more services"},
		{Text: "rollback", Description: "Revert changes to a service’s configuration"},
		{Text: "scale", Description: "Scale one or multiple replicated services"},
		{Text: "update", Description: "Update a service"},
	}
}

func containerListCompleter(all bool) []prompt.Suggest {
	suggestions := []prompt.Suggest{}
	ctx := context.Background()
	cList, _ := dockerClient.ContainerList(ctx, types.ContainerListOptions{All: all})

	for _, container := range cList {
		suggestions = append(suggestions, prompt.Suggest{Text: container.ID, Description: container.Image})
	}

	return suggestions
}

func main() {
	dockerClient, _ = docker.NewEnvClient()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	_, err := dockerClient.Info(ctx)
	if err != nil {
		fmt.Println("Couldn't check docker status please make sure docker is running.")
		fmt.Println(err)
		return
	}

run:
	dockerCommand := prompt.Input(">>> docker ",
		completer,
		prompt.OptionTitle("docker prompt"),
		prompt.OptionSelectedDescriptionTextColor(prompt.Turquoise),
		prompt.OptionInputTextColor(prompt.Fuchsia),
		prompt.OptionPrefixBackgroundColor(prompt.Cyan))

	splittedDockerCommands := strings.Split(dockerCommand, " ")
	if splittedDockerCommands[0] == "exit" {
		os.Exit(0)
	}

	var ps *exec.Cmd

	if splittedDockerCommands[0] == "clear" {
		ps = exec.Command("clear")
	} else {
		ps = exec.Command("docker", splittedDockerCommands...)
	}

	res, err := ps.Output()

	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(string(res))

	goto run
}
