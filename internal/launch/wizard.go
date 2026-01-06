package launch

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/selimozten/walgo/internal/config"
	"github.com/selimozten/walgo/internal/projects"
	"github.com/selimozten/walgo/internal/sui"
	"github.com/selimozten/walgo/internal/ui"
)

// sharedReader is a shared bufio.Reader to avoid creating multiple readers
var sharedReader *bufio.Reader

// getReader returns the shared reader, creating it if needed
func getReader() *bufio.Reader {
	if sharedReader == nil {
		sharedReader = bufio.NewReader(os.Stdin)
	}
	return sharedReader
}

// CloseReadline is a no-op now but kept for API compatibility
func CloseReadline() {
	// No cleanup needed for bufio.Reader
}

// ReadlineInput reads a line of input with the given prompt
// Uses simple bufio.Reader - terminal handles basic editing (backspace, etc.)
func ReadlineInput(prompt string) string {
	fmt.Print(prompt)
	reader := getReader()
	line, err := reader.ReadString('\n')
	if err != nil {
		return ""
	}
	return strings.TrimSpace(line)
}

// readlineInputWithDefault reads input with a default value shown
func readlineInputWithDefault(prompt, defaultVal string) string {
	var fullPrompt string
	if defaultVal != "" {
		fullPrompt = fmt.Sprintf("%s [%s]: ", prompt, defaultVal)
	} else {
		fullPrompt = prompt + ": "
	}

	result := ReadlineInput(fullPrompt)
	if result == "" {
		return defaultVal
	}
	return result
}

// SelectNetwork prompts the user to select a network
func SelectNetwork() (string, error) {
	// Get current active network using sui package
	currentNetwork, err := sui.GetActiveEnv()
	if err != nil {
		currentNetwork = "testnet"
	}

	fmt.Printf("Current network: %s\n\n", currentNetwork)
	fmt.Println("Available networks:")
	fmt.Println("  1) testnet  - For testing (1 epoch = 1 day)")
	fmt.Println("  2) mainnet  - For production (1 epoch = 2 weeks, requires SuiNS)")

	input := readlineInputWithDefault("\nSelect network", "1")

	if input == "" || input == "1" {
		return "testnet", nil
	} else if input == "2" {
		return "mainnet", nil
	}

	return "testnet", nil
}

// CheckWallet checks wallet and balance, returns (address, suiBalance, walBalance, error)
func CheckWallet(network string) (string, string, string, error) {

	icons := ui.GetIcons()

	// Switch to the selected network
	if err := sui.SwitchEnv(network); err != nil {
		return "", "", "", fmt.Errorf("failed to switch to %s: %w", network, err)
	}

	// Get active address
	activeAddr, err := sui.GetActiveAddress()
	if err != nil {
		return "", "", "", fmt.Errorf("failed to get wallet address: %w", err)
	}

	// Get all addresses
	addresses, err := sui.GetAddressList()
	if err != nil {
		return "", "", "", fmt.Errorf("failed to get addresses: %w", err)
	}

	// Loop until user makes a valid selection or chooses to go back
	for {
		// Display current wallet and balance
		fmt.Printf("\n%s Current Active Address:\n", icons.Info)
		fmt.Printf("   %s\n\n", activeAddr)

		suiBal, walBal := getBalance(activeAddr)
		fmt.Printf("%s Balance: %s SUI | %s WAL\n\n", icons.Arrow, suiBal, walBal)

		// Show wallet management menu
		fmt.Println("Wallet Options:")
		fmt.Println("  1) Use current address")
		fmt.Println("  2) Switch to another address")
		fmt.Println("  3) Create new address")
		fmt.Println("  4) Import existing address")
		fmt.Println("  b) Go back")

		input := readlineInputWithDefault("\nSelect", "1")

		switch input {
		case "", "1":
			// Use current address
			return activeAddr, suiBal, walBal, nil

		case "2":
			// Switch to another address
			addr, sui, wal, ok := switchAddressWithRetry(addresses, activeAddr)
			if ok {
				return addr, sui, wal, nil
			}
			// Not ok means user should retry, loop continues

		case "3":
			// Create new address
			return createNewAddress()

		case "4":
			// Import address
			return importAddress()

		case "b", "B":
			// Go back - return error to signal cancellation
			return "", "", "", fmt.Errorf("cancelled by user")

		default:
			fmt.Printf("%s Invalid option, please try again\n", icons.Warning)
		}
	}
}

// ProjectDetails holds all project metadata collected during launch
type ProjectDetails struct {
	Name        string
	Category    string
	Description string
	ImageURL    string
}

// DefaultWalgoLogoURL is the default logo used for Walrus Sites deployed with Walgo
const DefaultWalgoLogoURL = "https://cdn.jsdelivr.net/gh/selimozten/walgo@main/walgo-logo.svg"

// DefaultCategory is the default category for new projects
const DefaultCategory = "website"

// GetProjectDetails prompts for project name, category, and site metadata
// All fields have sensible defaults - user can just press Enter to accept them
func GetProjectDetails() (*ProjectDetails, error) {
	icons := ui.GetIcons()

	// Get current directory name as default project name
	cwd, err := os.Getwd()
	if err != nil {
		cwd = ""
	}
	defaultName := filepath.Base(cwd)
	if defaultName == "" || defaultName == "." || defaultName == "/" {
		defaultName = "my-walgo-site"
	}

	// Project name (also used as site_name for wallets/explorers)
	// Loop until a unique name is provided
	var name string
	for {
		name = readlineInputWithDefault("Project name", defaultName)
		if name == "" {
			name = defaultName
		}

		// Check if project name already exists
		pm, err := projects.NewManager()
		if err == nil {
			defer func() {
				_ = pm.Close()
			}()

			exists, err := pm.ProjectNameExists(name)
			if err == nil && exists {
				fmt.Printf("\n%s Project name '%s' already exists in your projects.\n", icons.Warning, name)
				fmt.Printf("%s Please choose a different name.\n", icons.Lightbulb)
				fmt.Println()
				defaultName = name // Keep the previous input as new default
				continue           // Ask again
			}
		}

		// If we get here, name is unique
		break
	}

	// Category
	category := readlineInputWithDefault("Category", DefaultCategory)
	if category == "" {
		category = DefaultCategory
	}

	// Description (auto-generated, can be changed later via walgo projects)
	defaultDesc := fmt.Sprintf("A %s deployed with Walgo to Walrus Sites", category)
	description := readlineInputWithDefault("Description", defaultDesc)
	if description == "" {
		description = defaultDesc
	}

	// Image URL (defaults to Walgo logo, can be changed later via walgo projects)
	imageURL := readlineInputWithDefault("Image URL", "Walgo logo")
	if imageURL == "" || imageURL == "Walgo logo" {
		imageURL = DefaultWalgoLogoURL
	}

	return &ProjectDetails{
		Name:        name,
		Category:    category,
		Description: description,
		ImageURL:    imageURL,
	}, nil
}

// SelectEpochs prompts for storage duration
func SelectEpochs(netConfig projects.NetworkConfig) (int, error) {
	fmt.Printf("Storage duration (epochs):\n")
	fmt.Printf("  • 1 epoch = %s\n", netConfig.EpochDuration)
	fmt.Printf("  • Maximum: %d epochs\n", netConfig.MaxEpochs)
	fmt.Println()

	var defaultEpochs string
	if netConfig.Name == "mainnet" {
		fmt.Println("Suggested durations:")
		fmt.Println("  • 2 epochs  = 1 month")
		fmt.Println("  • 6 epochs  = 3 months")
		fmt.Println("  • 26 epochs = 1 year")
		defaultEpochs = "5"
	} else {
		fmt.Println("Suggested durations:")
		fmt.Println("  • 7 epochs  = 1 week")
		fmt.Println("  • 30 epochs = 1 month")
		defaultEpochs = "1"
	}

	input := readlineInputWithDefault("\nEnter epochs", defaultEpochs)

	if input == "" {
		input = defaultEpochs
	}

	epochs, err := strconv.Atoi(input)
	if err != nil || epochs < 1 || epochs > netConfig.MaxEpochs {
		return 0, fmt.Errorf("invalid epochs (must be 1-%d)", netConfig.MaxEpochs)
	}

	return epochs, nil
}

// VerifySite verifies the site is built and ready
func VerifySite() (string, string, int64, error) {
	sitePath, err := os.Getwd()
	if err != nil {
		return "", "", 0, fmt.Errorf("failed to get current directory: %w", err)
	}

	walgoCfg, err := config.LoadConfigFrom(sitePath)
	if err != nil {
		return "", "", 0, fmt.Errorf("no walgo.yaml found - run 'walgo init' first")
	}

	publishDir := filepath.Join(sitePath, walgoCfg.HugoConfig.PublishDir)

	// Check if public directory exists
	_, err = os.Stat(publishDir)
	if os.IsNotExist(err) {
		return "", "", 0, fmt.Errorf("site not built - run 'walgo build' first")
	}

	// Calculate directory size
	var size int64
	filepath.Walk(publishDir, func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() {
			size += info.Size()
		}
		return nil
	})

	fmt.Printf("Site location: %s\n", publishDir)
	fmt.Printf("Site size: %.2f MB\n", float64(size)/(1024*1024))

	return sitePath, publishDir, size, nil
}

// Helper functions

func getBalance(address string) (string, string) {
	balance, err := sui.GetBalance()
	if err != nil {
		return "unknown", "unknown"
	}
	return fmt.Sprintf("%.2f", balance.SUI), fmt.Sprintf("%.2f", balance.WAL)
}

// switchAddressWithRetry attempts to switch address, returns (addr, suiBal, walBal, success)
// If success is false, the caller should retry/show menu again
func switchAddressWithRetry(addresses []string, currentAddr string) (string, string, string, bool) {
	icons := ui.GetIcons()
	if len(addresses) <= 1 {
		fmt.Printf("\n%s No other addresses available\n", icons.Warning)
		fmt.Printf("%s Try creating a new address or importing one\n\n", icons.Lightbulb)
		return "", "", "", false // Signal to retry
	}

	fmt.Printf("\n%s Available Addresses:\n", icons.Info)
	fmt.Println()

	for i, addr := range addresses {
		if addr == currentAddr {
			fmt.Printf("  %d) %s (current)\n", i+1, addr)
		} else {
			fmt.Printf("  %d) %s\n", i+1, addr)
		}
	}
	fmt.Println("  b) Go back")

	input := ReadlineInput("\nSelect address number: ")

	if input == "b" || input == "B" {
		return "", "", "", false // Go back to menu
	}

	selection, err := strconv.Atoi(input)
	if err != nil || selection < 1 || selection > len(addresses) {
		fmt.Printf("%s Invalid selection\n", icons.Warning)
		return "", "", "", false // Retry
	}

	selectedAddr := addresses[selection-1]

	// Switch to selected address
	if err := sui.SwitchAddress(selectedAddr); err != nil {
		fmt.Printf("%s Failed to switch address: %v\n", icons.Error, err)
		return "", "", "", false // Retry
	}

	fmt.Printf("\n%s Switched to: %s\n", icons.Check, selectedAddr)
	suiBal, walBal := getBalance(selectedAddr)
	fmt.Printf("%s Balance: %s SUI | %s WAL\n", icons.Arrow, suiBal, walBal)

	return selectedAddr, suiBal, walBal, true
}

func createNewAddress() (string, string, string, error) {
	icons := ui.GetIcons()
	fmt.Printf("\n%s Creating New Address\n", icons.Key)
	fmt.Println()
	fmt.Println("Key schemes:")
	fmt.Println("  1) ed25519   (recommended)")
	fmt.Println("  2) secp256k1")
	fmt.Println("  3) secp256r1")

	input := readlineInputWithDefault("\nSelect key scheme", "1")

	var keyScheme string
	switch input {
	case "", "1":
		keyScheme = "ed25519"
	case "2":
		keyScheme = "secp256k1"
	case "3":
		keyScheme = "secp256r1"
	default:
		keyScheme = "ed25519"
	}

	fmt.Printf("\n%s Creating new %s address...\n", icons.Spinner, keyScheme)

	result, err := sui.CreateAddressWithDetails(keyScheme, "")
	if err != nil {
		return "", "", "", fmt.Errorf("failed to create new address: %w", err)
	}

	newAddr := result.Address
	fmt.Printf("\n%s New address created!\n", icons.Check)
	fmt.Println()
	fmt.Printf("   Address: %s\n", newAddr)
	if result.Alias != "" {
		fmt.Printf("   Alias:   %s\n", result.Alias)
	}
	fmt.Println()

	// Display recovery phrase prominently
	if result.RecoveryPhrase != "" {
		phrase := result.RecoveryPhrase
		boxWidth := len(phrase) + 6 // 3 spaces padding on each side

		fmt.Printf("%s IMPORTANT: Save your recovery phrase!\n\n", icons.Warning)

		// Top border
		fmt.Print("   ┌")
		for i := 0; i < boxWidth; i++ {
			fmt.Print("─")
		}
		fmt.Println("┐")

		// Phrase
		fmt.Printf("   │   %s   │\n", phrase)

		// Bottom border
		fmt.Print("   └")
		for i := 0; i < boxWidth; i++ {
			fmt.Print("─")
		}
		fmt.Println("┘")

		fmt.Println()
		fmt.Println("   Store this phrase securely - it's the ONLY way to recover your wallet!")
		fmt.Println("   Never share it with anyone.")
		fmt.Println()
	}

	// Switch to the new address
	if err := sui.SwitchAddress(newAddr); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Failed to switch to new address: %v\n", err)
	}

	suiBal, walBal := getBalance(newAddr)
	fmt.Printf("%s Balance: %s SUI | %s WAL\n", icons.Arrow, suiBal, walBal)

	if suiBal == "unknown" || suiBal == "0.00" {
		fmt.Println()
		fmt.Printf("%s Your new address has no balance. You'll need to fund it:\n", icons.Lightbulb)
		env, _ := sui.GetActiveEnv()

		if strings.Contains(env, "testnet") {
			fmt.Printf("   • Testnet: https://faucet.sui.io/?address=%s\n", newAddr)
		} else {
			fmt.Println("   • Transfer SUI from another wallet")
			fmt.Println("   • Purchase SUI from an exchange")
		}
	}

	ReadlineInput("\nPress Enter to continue...")

	return newAddr, suiBal, walBal, nil
}

func importAddress() (string, string, string, error) {
	icons := ui.GetIcons()
	fmt.Printf("\n%s Import Existing Address\n", icons.Key)
	fmt.Println()
	fmt.Println("Import methods:")
	fmt.Println("  1) Private key (hex)")
	fmt.Println("  2) Mnemonic phrase")

	input := readlineInputWithDefault("\nSelect method", "1")

	var method sui.ImportMethod
	var keyScheme string
	var importInput string

	if input == "2" {
		method = sui.ImportFromMnemonic
		fmt.Printf("\n%s Enter your recovery phrase (12-24 words)\n", icons.Warning)
		keyScheme = readlineInputWithDefault("Key scheme", "ed25519")
		importInput = readlineInputWithDefault("Recovery phrase", "")
	} else {
		method = sui.ImportFromPrivateKey
		keyScheme = "ed25519"
		fmt.Printf("\n%s Enter your private key (suiprivkey1... or hex format)\n", icons.Warning)
		importInput = readlineInputWithDefault("Private key", "")
	}

	// Import the address
	fmt.Println()
	newAddr, err := sui.ImportAddressWithInput(method, keyScheme, importInput)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to import address: %w", err)
	}
	fmt.Printf("\n%s Address imported: %s\n", icons.Check, newAddr)

	suiBal, walBal := getBalance(newAddr)
	fmt.Printf("%s Balance: %s SUI | %s WAL\n", icons.Arrow, suiBal, walBal)

	return newAddr, suiBal, walBal, nil
}
