package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"mindful/src/models"
)

// promptUser prompts the user for conflict resolution with detailed information
func promptUser(conflicts []*models.FileConflict, toolName string) (models.ConflictResolution, error) {
	fmt.Printf("\n⚠️  Found %d conflict(s) for %s:\n", len(conflicts), toolName)

	// Display detailed conflict information
	for i, conflict := range conflicts {
		fmt.Printf("\n%d. File: %s (%s)\n", i+1, conflict.FilePath, conflict.FileType)
		fmt.Printf("   Existing hash: %s\n", conflict.ExistingHash)
		fmt.Printf("   New hash: %s\n", conflict.NewHash)
		fmt.Printf("   Changes: %s\n", conflict.Diff)
	}

	fmt.Printf("\nHow would you like to proceed?\n")
	fmt.Printf("  [c] Continue - overwrite this conflict and continue\n")
	fmt.Printf("  [a] Continue All - overwrite all conflicts without further prompting\n")
	fmt.Printf("  [s] Stop - halt the operation (default)\n")
	fmt.Printf("\nChoice [c/a/s]: ")

	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return models.Stop, fmt.Errorf("failed to read user input: %w", err)
	}

	choice := strings.ToLower(strings.TrimSpace(input))
	switch choice {
	case "c", "continue":
		return models.Continue, nil
	case "a", "all", "continue all":
		return models.ContinueAll, nil
	case "s", "stop", "":
		return models.Stop, nil
	default:
		fmt.Printf("Invalid choice '%s', defaulting to Stop\n", choice)
		return models.Stop, nil
	}
}

// promptSingleConflict prompts the user for a single conflict resolution
func promptSingleConflict(conflict *models.FileConflict, toolName string, conflictIndex, totalConflicts int) (models.ConflictResolution, error) {
	fmt.Printf("\n⚠️  Conflict %d of %d for %s:\n", conflictIndex+1, totalConflicts, toolName)
	fmt.Printf("   File: %s (%s)\n", conflict.FilePath, conflict.FileType)
	fmt.Printf("   Existing hash: %s\n", conflict.ExistingHash)
	fmt.Printf("   New hash: %s\n", conflict.NewHash)
	fmt.Printf("   Changes: %s\n", conflict.Diff)

	fmt.Printf("\nHow would you like to proceed?\n")
	fmt.Printf("  [c] Continue - overwrite this conflict and continue\n")
	fmt.Printf("  [a] Continue All - overwrite all remaining conflicts without further prompting\n")
	fmt.Printf("  [s] Stop - halt the operation (default)\n")
	fmt.Printf("\nChoice [c/a/s]: ")

	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return models.Stop, fmt.Errorf("failed to read user input: %w", err)
	}

	choice := strings.ToLower(strings.TrimSpace(input))
	switch choice {
	case "c", "continue":
		return models.Continue, nil
	case "a", "all", "continue all":
		return models.ContinueAll, nil
	case "s", "stop", "":
		return models.Stop, nil
	default:
		fmt.Printf("Invalid choice '%s', defaulting to Stop\n", choice)
		return models.Stop, nil
	}
}
