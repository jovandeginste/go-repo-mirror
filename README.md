# Installation

```
go get github.com/jovandeginste/go-repo-mirror
```

# Usage

```
usage: go-repo-mirror [<flags>] <repo-url> <destination-folder>

Flags:
      --help                     Show context-sensitive help (also try --help-long and --help-man).
  -v, --verbose=1                Verbosity level (0 to silence).
  -l, --log-file=LOG-FILE        File to write logs to (logs still go to stdout).
  -m, --metadata-only            Only download repository metadata.
  -d, --data-only                Only download repository data.
  -c, --concurrent-downloads=10  Number of concurrent downloads.
      --size-check               Don't verify file hash.
      --cert=CERT                Client certificate file (PEM).
      --key=KEY                  Client private key file (PEM).
      --insecure-tls             Disable TLS check for server.

Args:
  <repo-url>            Remote URL to mirror the repository from.
  <destination-folder>  Local folder to mirror the repository to.
```

This mirrors a remote yum-compatible repository (starting from `$URL/repodata/repomd.xml`) to a local
directory, keeping the file structure. Default behaviour is to download files that don't exist or don't
have the same checksum as the metadata suggests. You can speed things up by only checking file size.

You can limit to only data or only metadata.

Works for CentOS and RedHat repositories. For RedHat, you need official entitlements (cert and key).

Example usage:

```
go-repo-mirror -v http://mirror.centos.org/centos/7/os/x86_64 /var/repo/centos/7/os/x86_64/
```
