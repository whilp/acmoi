/*

Watch calls various helpers on put events emitted by acme. When it sees such an event, it attempts to invoke each of the following external executables until it encounters an error:

- `acme-format`
- `acme-check`
- `acme-build`
- `acme-test`

First, Watch determines the root of the project containing the file that has just been written by invoking `acme-root` with no arguments. Then, it invokes each command in sequence until it encounters an error, passing the path to the newly written file relative to the project root. The stdout and stderr of the commands are routed to the project root's +Errors window. `acme-format` is expected to rewrite the file on disk; if it succeeds, the window containing the file will be refreshed. Any command may return a non-zero status to stop further execution.

*/

package main // import "github.com/whilp/acmoi/cmd/Watch"
