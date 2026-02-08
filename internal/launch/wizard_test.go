package launch

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/selimozten/walgo/internal/projects"
)

// setTestReader sets a custom reader for testing input
func setTestReader(input string) {
	sharedReader = bufio.NewReader(strings.NewReader(input))
}

// resetReader resets the shared reader to nil so getReader() creates a new one
func resetReader() {
	sharedReader = nil
}

// ====================
// Tests for getReader()
// ====================

func TestGetReader(t *testing.T) {
	t.Run("creates reader if nil", func(t *testing.T) {
		resetReader()
		setTestReader("test\n")
		reader := getReader()
		if reader == nil {
			t.Error("getReader should return a non-nil reader")
		}
	})

	t.Run("returns same reader on subsequent calls", func(t *testing.T) {
		setTestReader("test\n")
		reader1 := getReader()
		reader2 := getReader()
		if reader1 != reader2 {
			t.Error("getReader should return the same reader instance")
		}
	})

	t.Run("nil sharedReader creates new reader", func(t *testing.T) {
		sharedReader = nil
		setTestReader("data\n")
		reader := getReader()
		if reader == nil {
			t.Error("Expected non-nil reader after setting test reader")
		}
	})
}

// ====================
// Tests for CloseReadline()
// ====================

func TestCloseReadline(t *testing.T) {
	// This is a no-op function, just ensure it doesn't panic
	t.Run("does not panic", func(t *testing.T) {
		CloseReadline()
	})

	t.Run("can be called multiple times", func(t *testing.T) {
		CloseReadline()
		CloseReadline()
		CloseReadline()
	})
}

// ====================
// Tests for ReadlineInput()
// ====================

func TestReadlineInput(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"simple input", "hello\n", "hello"},
		{"input with whitespace", "  hello world  \n", "hello world"},
		{"empty input", "\n", ""},
		{"input with special characters", "test@123!#\n", "test@123!#"},
		{"input with tabs", "\t\thello\t\t\n", "hello"},
		{"numeric input", "12345\n", "12345"},
		{"input with hyphens", "my-test-input\n", "my-test-input"},
		{"input with underscores", "my_test_input\n", "my_test_input"},
		{"url-like input", "https://example.com\n", "https://example.com"},
		{"path-like input", "/path/to/file\n", "/path/to/file"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setTestReader(tt.input)
			result := ReadlineInput("prompt: ")
			if result != tt.expected {
				t.Errorf("ReadlineInput() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestReadlineInputEmptyOnError(t *testing.T) {
	setTestReader("")
	result := ReadlineInput("prompt: ")
	if result != "" {
		t.Errorf("ReadlineInput() on EOF = %q, want empty string", result)
	}
}

func TestReadlineInputWithMultipleLines(t *testing.T) {
	input := "first\nsecond\nthird\n"
	setTestReader(input)

	result1 := ReadlineInput("prompt1: ")
	if result1 != "first" {
		t.Errorf("First read = %q, want 'first'", result1)
	}

	result2 := ReadlineInput("prompt2: ")
	if result2 != "second" {
		t.Errorf("Second read = %q, want 'second'", result2)
	}

	result3 := ReadlineInput("prompt3: ")
	if result3 != "third" {
		t.Errorf("Third read = %q, want 'third'", result3)
	}
}

// ====================
// Tests for readlineInputWithDefault()
// ====================

func TestReadlineInputWithDefault(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		prompt     string
		defaultVal string
		expected   string
	}{
		{"user provides input", "custom\n", "Test prompt", "default", "custom"},
		{"user accepts default", "\n", "Test prompt", "default", "default"},
		{"empty default and empty input", "\n", "Test prompt", "", ""},
		{"whitespace only returns default", "   \n", "Test prompt", "default", "default"},
		{"default with special chars", "\n", "Test", "test@123", "test@123"},
		{"override special char default", "new\n", "Test", "test@123", "new"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setTestReader(tt.input)
			result := readlineInputWithDefault(tt.prompt, tt.defaultVal)
			if result != tt.expected {
				t.Errorf("readlineInputWithDefault() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestReadlineInputWithDefaultPromptFormat(t *testing.T) {
	t.Run("with default value shows brackets", func(t *testing.T) {
		setTestReader("\n")
		// The function adds [default]: suffix when default is provided
		result := readlineInputWithDefault("Enter value", "mydefault")
		if result != "mydefault" {
			t.Errorf("Expected default value 'mydefault', got %q", result)
		}
	})

	t.Run("without default value shows colon only", func(t *testing.T) {
		setTestReader("myinput\n")
		result := readlineInputWithDefault("Enter value", "")
		if result != "myinput" {
			t.Errorf("Expected 'myinput', got %q", result)
		}
	})
}

// ====================
// Tests for SelectEpochs()
// ====================

func TestSelectEpochs(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		network  string
		wantErr  bool
		expected int
	}{
		{"valid epoch 1 testnet", "1\n", "testnet", false, 1},
		{"valid epoch at max", "53\n", "testnet", false, 53},
		{"epoch exceeds max", "54\n", "testnet", true, 0},
		{"epoch zero invalid", "0\n", "testnet", true, 0},
		{"negative epoch invalid", "-1\n", "testnet", true, 0},
		{"non-numeric invalid", "abc\n", "testnet", true, 0},
		{"mainnet valid epoch", "5\n", "mainnet", false, 5},
		{"mainnet default", "\n", "mainnet", false, 5},
		{"testnet default", "\n", "testnet", false, 1},
		{"valid mid-range epoch", "25\n", "testnet", false, 25},
		{"valid high epoch", "52\n", "testnet", false, 52},
		{"decimal input invalid", "1.5\n", "testnet", true, 0},
		{"spaces in input", "  10  \n", "testnet", false, 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setTestReader(tt.input)
			netConfig := projects.GetNetworkConfig(tt.network)
			epochs, err := SelectEpochs(netConfig)

			if (err != nil) != tt.wantErr {
				t.Errorf("SelectEpochs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && epochs != tt.expected {
				t.Errorf("SelectEpochs() = %d, want %d", epochs, tt.expected)
			}
		})
	}
}

func TestSelectEpochsNetworkConfig(t *testing.T) {
	t.Run("testnet config", func(t *testing.T) {
		netConfig := projects.GetNetworkConfig("testnet")
		if netConfig.Name != "testnet" {
			t.Errorf("GetNetworkConfig(testnet).Name = %q, want 'testnet'", netConfig.Name)
		}
		if netConfig.MaxEpochs != 53 {
			t.Errorf("GetNetworkConfig(testnet).MaxEpochs = %d, want 53", netConfig.MaxEpochs)
		}
		if netConfig.EpochDuration != "1 day" {
			t.Errorf("GetNetworkConfig(testnet).EpochDuration = %q, want '1 day'", netConfig.EpochDuration)
		}
	})

	t.Run("mainnet config", func(t *testing.T) {
		netConfig := projects.GetNetworkConfig("mainnet")
		if netConfig.Name != "mainnet" {
			t.Errorf("GetNetworkConfig(mainnet).Name = %q, want 'mainnet'", netConfig.Name)
		}
		if netConfig.MaxEpochs != 53 {
			t.Errorf("GetNetworkConfig(mainnet).MaxEpochs = %d, want 53", netConfig.MaxEpochs)
		}
		if netConfig.EpochDuration != "2 weeks" {
			t.Errorf("GetNetworkConfig(mainnet).EpochDuration = %q, want '2 weeks'", netConfig.EpochDuration)
		}
	})

	t.Run("unknown network defaults to testnet", func(t *testing.T) {
		netConfig := projects.GetNetworkConfig("unknown")
		if netConfig.Name != "testnet" {
			t.Errorf("GetNetworkConfig(unknown).Name = %q, want 'testnet'", netConfig.Name)
		}
	})
}

// ====================
// Tests for VerifySite()
// ====================

func TestVerifySiteRequiresConfig(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "walgo-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current dir: %v", err)
	}
	defer func() { _ = os.Chdir(originalDir) }()

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to chdir: %v", err)
	}

	_, _, _, err = VerifySite()
	if err == nil {
		t.Error("VerifySite() should fail without walgo.yaml")
	}
	if !strings.Contains(err.Error(), "walgo.yaml") && !strings.Contains(err.Error(), "walgo init") {
		t.Errorf("Error should mention walgo.yaml, got: %v", err)
	}
}

func TestVerifySiteDirectorySizeCalculation(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "walgo-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	publicDir := filepath.Join(tmpDir, "public")
	if err := os.MkdirAll(publicDir, 0755); err != nil {
		t.Fatalf("Failed to create public dir: %v", err)
	}

	files := map[string]int{
		"index.html": 100,
		"style.css":  200,
		"app.js":     300,
	}

	for name, size := range files {
		content := make([]byte, size)
		if err := os.WriteFile(filepath.Join(publicDir, name), content, 0644); err != nil {
			t.Fatalf("Failed to create file %s: %v", name, err)
		}
	}

	var size int64
	err = filepath.Walk(publicDir, func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	if err != nil {
		t.Fatalf("Failed to walk directory: %v", err)
	}

	expectedSize := int64(100 + 200 + 300)
	if size != expectedSize {
		t.Errorf("Directory size = %d, want %d", size, expectedSize)
	}
}

// ====================
// Tests for ProjectDetails struct and constants
// ====================

func TestProjectDetailsDefaults(t *testing.T) {
	t.Run("DefaultWalgoLogoURL is set", func(t *testing.T) {
		if DefaultWalgoLogoURL == "" {
			t.Error("DefaultWalgoLogoURL should not be empty")
		}
		if !strings.HasPrefix(DefaultWalgoLogoURL, "https://") {
			t.Error("DefaultWalgoLogoURL should be a valid HTTPS URL")
		}
		// Check for GitHub hosting - either direct github.com or jsdelivr CDN (uses "gh" for GitHub)
		if !strings.Contains(DefaultWalgoLogoURL, "github") && !strings.Contains(DefaultWalgoLogoURL, "jsdelivr.net/gh/") {
			t.Error("DefaultWalgoLogoURL should be hosted on GitHub or jsdelivr CDN")
		}
	})

	t.Run("DefaultCategory is website", func(t *testing.T) {
		if DefaultCategory != "website" {
			t.Errorf("DefaultCategory = %q, want %q", DefaultCategory, "website")
		}
	})
}

func TestProjectDetailsStruct(t *testing.T) {
	pd := &ProjectDetails{
		Name:        "test-project",
		Category:    "blog",
		Description: "A test project",
		ImageURL:    "https://example.com/image.png",
	}

	if pd.Name != "test-project" {
		t.Errorf("ProjectDetails.Name = %q, want %q", pd.Name, "test-project")
	}
	if pd.Category != "blog" {
		t.Errorf("ProjectDetails.Category = %q, want %q", pd.Category, "blog")
	}
	if pd.Description != "A test project" {
		t.Errorf("ProjectDetails.Description = %q, want %q", pd.Description, "A test project")
	}
	if pd.ImageURL != "https://example.com/image.png" {
		t.Errorf("ProjectDetails.ImageURL = %q, want %q", pd.ImageURL, "https://example.com/image.png")
	}
}

func TestProjectDetailsEmptyFields(t *testing.T) {
	pd := &ProjectDetails{}
	if pd.Name != "" {
		t.Error("Empty ProjectDetails.Name should be empty string")
	}
	if pd.Category != "" {
		t.Error("Empty ProjectDetails.Category should be empty string")
	}
}

// ====================
// Tests for network selection logic
// ====================

func TestNetworkSelectionLogic(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"select testnet with 1", "1\n", "testnet"},
		{"select mainnet with 2", "2\n", "mainnet"},
		{"empty defaults to testnet", "\n", "testnet"},
		{"invalid defaults to testnet", "invalid\n", "testnet"},
		{"number 3 defaults to testnet", "3\n", "testnet"},
		{"number 0 defaults to testnet", "0\n", "testnet"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setTestReader(tt.input)
			input := readlineInputWithDefault("\nSelect network", "1")

			var result string
			if input == "" || input == "1" {
				result = "testnet"
			} else if input == "2" {
				result = "mainnet"
			} else {
				result = "testnet"
			}

			if result != tt.expected {
				t.Errorf("network selection = %q, want %q", result, tt.expected)
			}
		})
	}
}

// ====================
// Tests for balance formatting
// ====================

func TestBalanceFormatting(t *testing.T) {
	tests := []struct {
		name     string
		amount   float64
		expected string
	}{
		{"zero balance", 0.00, "0.00"},
		{"normal balance", 10.50, "10.50"},
		{"large balance", 1000000.99, "1000000.99"},
		{"small decimal", 0.01, "0.01"},
		{"very small", 0.001, "0.00"},
		{"whole number", 100.00, "100.00"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatted := fmt.Sprintf("%.2f", tt.amount)
			if formatted != tt.expected {
				t.Errorf("formatBalance() = %q, want %q", formatted, tt.expected)
			}
		})
	}
}

// ====================
// Tests for key scheme selection logic
// ====================

func TestKeySchemeSelection(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"1", "ed25519"},
		{"2", "secp256k1"},
		{"3", "secp256r1"},
		{"", "ed25519"},
		{"4", "ed25519"},
		{"invalid", "ed25519"},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("input_%s", tt.input), func(t *testing.T) {
			var keyScheme string
			switch tt.input {
			case "", "1":
				keyScheme = "ed25519"
			case "2":
				keyScheme = "secp256k1"
			case "3":
				keyScheme = "secp256r1"
			default:
				keyScheme = "ed25519"
			}

			if keyScheme != tt.expected {
				t.Errorf("keyScheme = %q, want %q", keyScheme, tt.expected)
			}
		})
	}
}

// ====================
// Tests for wallet option selection logic
// ====================

func TestWalletOptionSelection(t *testing.T) {
	tests := []struct {
		input  string
		action string
	}{
		{"1", "use_current"},
		{"", "use_current"},
		{"2", "switch"},
		{"3", "create"},
		{"4", "import"},
		{"b", "back"},
		{"B", "back"},
		{"invalid", "invalid"},
		{"5", "invalid"},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("input_%s", tt.input), func(t *testing.T) {
			var action string
			switch tt.input {
			case "", "1":
				action = "use_current"
			case "2":
				action = "switch"
			case "3":
				action = "create"
			case "4":
				action = "import"
			case "b", "B":
				action = "back"
			default:
				action = "invalid"
			}

			if action != tt.action {
				t.Errorf("action = %q, want %q", action, tt.action)
			}
		})
	}
}

// ====================
// Tests for address switch validation
// ====================

func TestSwitchAddressValidation(t *testing.T) {
	tests := []struct {
		name          string
		addresses     []string
		input         string
		expectSuccess bool
	}{
		{"single address", []string{"0x123"}, "1", false},
		{"multiple - valid", []string{"0x123", "0x456"}, "2", true},
		{"multiple - first", []string{"0x123", "0x456"}, "1", true},
		{"out of range", []string{"0x123", "0x456"}, "3", false},
		{"go back b", []string{"0x123", "0x456"}, "b", false},
		{"go back B", []string{"0x123", "0x456"}, "B", false},
		{"non-numeric", []string{"0x123", "0x456"}, "abc", false},
		{"zero selection", []string{"0x123", "0x456"}, "0", false},
		{"negative selection", []string{"0x123", "0x456"}, "-1", false},
		{"empty addresses", []string{}, "1", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			success := validateAddressSwitch(tt.addresses, tt.input)
			if success != tt.expectSuccess {
				t.Errorf("validateAddressSwitch() = %v, want %v", success, tt.expectSuccess)
			}
		})
	}
}

func validateAddressSwitch(addresses []string, input string) bool {
	if len(addresses) <= 1 {
		return false
	}
	if input == "b" || input == "B" {
		return false
	}

	selection, err := strconv.Atoi(input)
	if err != nil {
		return false
	}

	if selection < 1 || selection > len(addresses) {
		return false
	}

	return true
}

// ====================
// Tests for import method selection
// ====================

func TestImportMethodSelection(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"1", "key"},
		{"", "key"},
		{"2", "mnemonic"},
		{"3", "key"},
		{"invalid", "key"},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("input_%s", tt.input), func(t *testing.T) {
			var method string
			if tt.input == "2" {
				method = "mnemonic"
			} else {
				method = "key"
			}

			if method != tt.expected {
				t.Errorf("import method = %q, want %q", method, tt.expected)
			}
		})
	}
}

// ====================
// Tests for project name derivation
// ====================

func TestDefaultProjectName(t *testing.T) {
	tests := []struct {
		cwd      string
		expected string
	}{
		{"/home/user/my-blog", "my-blog"},
		{"/", "my-walgo-site"},
		{".", "my-walgo-site"},
		{"/Users/test/projects/awesome-site", "awesome-site"},
		{"/path/to/my-project-123", "my-project-123"},
		{"", "my-walgo-site"},
	}

	for _, tt := range tests {
		t.Run(tt.cwd, func(t *testing.T) {
			defaultName := deriveDefaultProjectName(tt.cwd)
			if defaultName != tt.expected {
				t.Errorf("deriveDefaultProjectName(%q) = %q, want %q", tt.cwd, defaultName, tt.expected)
			}
		})
	}
}

func deriveDefaultProjectName(cwd string) string {
	defaultName := filepath.Base(cwd)
	if defaultName == "" || defaultName == "." || defaultName == "/" {
		defaultName = "my-walgo-site"
	}
	return defaultName
}

// ====================
// Tests for image URL defaulting
// ====================

func TestImageURLDefault(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"", DefaultWalgoLogoURL},
		{"Walgo logo", DefaultWalgoLogoURL},
		{"https://example.com/logo.png", "https://example.com/logo.png"},
		{"custom-url", "custom-url"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			imageURL := tt.input
			if imageURL == "" || imageURL == "Walgo logo" {
				imageURL = DefaultWalgoLogoURL
			}
			if imageURL != tt.expected {
				t.Errorf("imageURL = %q, want %q", imageURL, tt.expected)
			}
		})
	}
}

// ====================
// Tests for category defaulting
// ====================

func TestCategoryDefault(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"", DefaultCategory},
		{"blog", "blog"},
		{"portfolio", "portfolio"},
		{"website", "website"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			category := tt.input
			if category == "" {
				category = DefaultCategory
			}
			if category != tt.expected {
				t.Errorf("category = %q, want %q", category, tt.expected)
			}
		})
	}
}

// ====================
// Tests for description generation
// ====================

func TestDescriptionDefault(t *testing.T) {
	tests := []struct {
		category string
		expected string
	}{
		{"website", "A website deployed with Walgo to Walrus Sites"},
		{"blog", "A blog deployed with Walgo to Walrus Sites"},
		{"portfolio", "A portfolio deployed with Walgo to Walrus Sites"},
		{"app", "A app deployed with Walgo to Walrus Sites"},
	}

	for _, tt := range tests {
		t.Run(tt.category, func(t *testing.T) {
			defaultDesc := fmt.Sprintf("A %s deployed with Walgo to Walrus Sites", tt.category)
			if defaultDesc != tt.expected {
				t.Errorf("description = %q, want %q", defaultDesc, tt.expected)
			}
		})
	}
}

// ====================
// Tests for epoch validation
// ====================

func TestEpochValidationRange(t *testing.T) {
	maxEpochs := 53

	tests := []struct {
		input   int
		isValid bool
	}{
		{-10, false},
		{-1, false},
		{0, false},
		{1, true},
		{26, true},
		{53, true},
		{54, false},
		{100, false},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("epoch_%d", tt.input), func(t *testing.T) {
			isValid := tt.input >= 1 && tt.input <= maxEpochs
			if isValid != tt.isValid {
				t.Errorf("Epoch %d validation = %v, want %v", tt.input, isValid, tt.isValid)
			}
		})
	}
}

// ====================
// Tests for network epoch defaults
// ====================

func TestNetworkEpochDefaults(t *testing.T) {
	tests := []struct {
		network       string
		defaultEpochs string
	}{
		{"mainnet", "5"},
		{"testnet", "1"},
		{"unknown", "1"},
	}

	for _, tt := range tests {
		t.Run(tt.network, func(t *testing.T) {
			var defaultEpochs string
			if tt.network == "mainnet" {
				defaultEpochs = "5"
			} else {
				defaultEpochs = "1"
			}
			if defaultEpochs != tt.defaultEpochs {
				t.Errorf("defaultEpochs = %q, want %q", defaultEpochs, tt.defaultEpochs)
			}
		})
	}
}

// ====================
// Tests for storage duration calculation
// ====================

func TestStorageDurationCalculation(t *testing.T) {
	tests := []struct {
		epochs  int
		network string
	}{
		{1, "testnet"},
		{7, "testnet"},
		{14, "testnet"},
		{30, "testnet"},
		{1, "mainnet"},
		{2, "mainnet"},
		{4, "mainnet"},
		{26, "mainnet"},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s_%d", tt.network, tt.epochs), func(t *testing.T) {
			result := projects.CalculateStorageDuration(tt.epochs, tt.network)
			if result == "" {
				t.Errorf("CalculateStorageDuration(%d, %q) returned empty string", tt.epochs, tt.network)
			}
		})
	}
}

// ====================
// Tests for recovery phrase box width
// ====================

func TestRecoveryPhraseBoxWidth(t *testing.T) {
	tests := []struct {
		name          string
		phrase        string
		expectedWidth int
	}{
		{"short phrase", "word1 word2 word3", len("word1 word2 word3") + 6},
		{"12 words", "a b c d e f g h i j k l", len("a b c d e f g h i j k l") + 6},
		{"single word", "single", len("single") + 6},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			boxWidth := len(tt.phrase) + 6
			if boxWidth != tt.expectedWidth {
				t.Errorf("boxWidth = %d, want %d", boxWidth, tt.expectedWidth)
			}
		})
	}
}

// ====================
// Tests for faucet URL generation
// ====================

func TestFaucetURLGeneration(t *testing.T) {
	addresses := []string{
		"0x1234567890abcdef1234567890abcdef12345678",
		"0xabcdef1234567890abcdef1234567890abcdef12",
		"0x0000000000000000000000000000000000000000",
	}

	for _, addr := range addresses {
		t.Run(addr[:10]+"...", func(t *testing.T) {
			faucetURL := fmt.Sprintf("https://faucet.sui.io/?address=%s", addr)
			if !strings.HasPrefix(faucetURL, "https://faucet.sui.io/?address=") {
				t.Error("Faucet URL should start with proper prefix")
			}
			if !strings.Contains(faucetURL, addr) {
				t.Error("Faucet URL should contain the address")
			}
		})
	}
}

// ====================
// Tests for size formatting
// ====================

func TestSizeFormatting(t *testing.T) {
	tests := []struct {
		bytes    int64
		expected float64
	}{
		{0, 0.0},
		{1024 * 1024, 1.0},
		{512 * 1024, 0.5},
		{10 * 1024 * 1024, 10.0},
		{100 * 1024, 0.09765625},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%d_bytes", tt.bytes), func(t *testing.T) {
			mb := float64(tt.bytes) / (1024 * 1024)
			if mb != tt.expected {
				t.Errorf("Size in MB = %f, want %f", mb, tt.expected)
			}
		})
	}
}

func TestSizeFormattingString(t *testing.T) {
	tests := []struct {
		bytes    int64
		expected string
	}{
		{0, "0.00"},
		{1024 * 1024, "1.00"},
		{1536 * 1024, "1.50"},
		{2048 * 1024, "2.00"},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%d_bytes", tt.bytes), func(t *testing.T) {
			formatted := fmt.Sprintf("%.2f", float64(tt.bytes)/(1024*1024))
			if formatted != tt.expected {
				t.Errorf("Size formatting = %q, want %q", formatted, tt.expected)
			}
		})
	}
}

// ====================
// Tests for address list handling
// ====================

func TestAddressListHandling(t *testing.T) {
	t.Run("empty list", func(t *testing.T) {
		addresses := []string{}
		if len(addresses) > 1 {
			t.Error("Empty list should not have len > 1")
		}
	})

	t.Run("single address", func(t *testing.T) {
		addresses := []string{"0x123"}
		if len(addresses) > 1 {
			t.Error("Single address list should not have len > 1")
		}
	})

	t.Run("multiple addresses", func(t *testing.T) {
		addresses := []string{"0x123", "0x456", "0x789"}
		if len(addresses) <= 1 {
			t.Error("Multiple addresses should have len > 1")
		}
	})
}

func TestAddressListDisplay(t *testing.T) {
	addresses := []string{"0x123abc", "0x456def", "0x789ghi"}
	currentAddr := "0x123abc"

	var output strings.Builder
	for i, addr := range addresses {
		if addr == currentAddr {
			output.WriteString(fmt.Sprintf("  %d) %s (current)\n", i+1, addr))
		} else {
			output.WriteString(fmt.Sprintf("  %d) %s\n", i+1, addr))
		}
	}

	result := output.String()

	if !strings.Contains(result, "(current)") {
		t.Error("Current address should be marked")
	}
	if !strings.Contains(result, "0x123abc (current)") {
		t.Error("First address should be marked as current")
	}
	if strings.Contains(result, "0x456def (current)") {
		t.Error("Second address should not be marked as current")
	}
}

// ====================
// Tests for concurrent access
// ====================

func TestConcurrentReaderAccess(t *testing.T) {
	setTestReader("test\ntest\ntest\ntest\ntest\n")

	done := make(chan bool)
	for i := 0; i < 5; i++ {
		go func() {
			reader := getReader()
			_ = reader
			done <- true
		}()
	}

	for i := 0; i < 5; i++ {
		<-done
	}
}

// ====================
// Benchmark tests
// ====================

func BenchmarkReadlineInput(b *testing.B) {
	input := "test input line\n"
	for i := 0; i < b.N; i++ {
		setTestReader(input)
		ReadlineInput("prompt: ")
	}
}

func BenchmarkDeriveDefaultProjectName(b *testing.B) {
	cwd := "/home/user/my-awesome-project"
	for i := 0; i < b.N; i++ {
		deriveDefaultProjectName(cwd)
	}
}

func BenchmarkSelectEpochs(b *testing.B) {
	netConfig := projects.GetNetworkConfig("testnet")
	for i := 0; i < b.N; i++ {
		setTestReader("10\n")
		_, _ = SelectEpochs(netConfig)
	}
}

func BenchmarkAddressValidation(b *testing.B) {
	addresses := []string{"0x123", "0x456", "0x789", "0xabc", "0xdef"}
	for i := 0; i < b.N; i++ {
		validateAddressSwitch(addresses, "3")
	}
}

func BenchmarkReadlineInputWithDefault(b *testing.B) {
	for i := 0; i < b.N; i++ {
		setTestReader("input\n")
		readlineInputWithDefault("prompt", "default")
	}
}
