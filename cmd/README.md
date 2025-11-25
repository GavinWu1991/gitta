# CLI Adapters (`cmd/`)

**Purpose**: Thin command-line adapters that translate user input (flags, arguments) into domain requests.

**Responsibilities**:
- Parse Cobra commands and flags
- Wire global flags (`--json`, `--log-level`)
- Delegate business logic to `internal/services` via interfaces
- Format output (human-readable or JSON)

**Allowed Dependencies**:
- `internal/core` (domain interfaces)
- `internal/services` (service implementations)
- `pkg/config`, `pkg/logging` (shared utilities)
- Cobra, standard library

**Forbidden Dependencies**:
- `infra/` (use via service interfaces)
- `ui/` (separate adapter layer)

## Adding New Commands

To add a new Cobra command without violating hexagonal boundaries:

1. **Create command file**: `cmd/gitta/<command>.go`
   ```go
   package main
   
   import (
       "github.com/gavin/gitta/internal/services"
       "github.com/spf13/cobra"
   )
   
   var myCmd = &cobra.Command{
       Use:   "mycommand",
       Short: "Brief description",
       Long:  "Detailed description",
       RunE: func(cmd *cobra.Command, args []string) error {
           // Inject service via constructor (not global)
           svc := services.NewMyService() // or inject via root
           return svc.DoSomething(args)
       },
   }
   ```

2. **Register in root.go**: Add to `init()` function:
   ```go
   func init() {
       rootCmd.AddCommand(versionCmd)
       rootCmd.AddCommand(myCmd) // Add your command
   }
   ```

3. **Dependency Injection Pattern**:
   - Services should be created in `RunE`, not as package-level variables
   - Pass `context.Context` from `cmd.Context()` to services
   - Services implement interfaces from `internal/core`
   - Never import `infra/` directly - use service interfaces

4. **Update Contracts**: Add command specification to `specs/*/contracts/cli.md`

**Example**: `cmd/gitta/root.go` wires flags and calls `services.TaskService.List()`.

