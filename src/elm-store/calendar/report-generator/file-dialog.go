package report_generator

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// SignatureImageSelectedMsg is dispatched when a signature image is selected
type SignatureImageSelectedMsg struct {
	ImagePath string
}

// OpenImageFileDialog opens a native file dialog for selecting an image
func OpenImageFileDialog() tea.Cmd {
	return func() tea.Msg {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			homeDir = "/"
		}

		var cmd *exec.Cmd

		switch {
		case commandExists("zenity"):
			cmd = exec.Command("zenity", "--file-selection",
				"--title=Select Signature Image",
				"--file-filter=Images | *.png *.jpg *.jpeg",
				"--filename="+homeDir+"/")
		case commandExists("kdialog"):
			cmd = exec.Command("kdialog", "--getopenfilename", homeDir, "*.png *.jpg *.jpeg")
		case commandExists("osascript"):
			script := `
				set imageFile to choose file with prompt "Select Signature Image" of type {"public.image"}
				return POSIX path of imageFile
			`
			cmd = exec.Command("osascript", "-e", script)
		default:
			return SignatureImageSelectedMsg{ImagePath: ""}
		}

		output, err := cmd.Output()
		if err != nil {
			return SignatureImageSelectedMsg{ImagePath: ""}
		}

		imagePath := strings.TrimSpace(string(output))
		return SignatureImageSelectedMsg{ImagePath: imagePath}
	}
}

// OpenCSVSaveDialog opens a save dialog for CSV files
func OpenCSVSaveDialog(sourceFile string) (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	defaultFileName := filepath.Base(sourceFile)
	defaultPath := filepath.Join(homeDir, defaultFileName)

	var cmd *exec.Cmd

	switch {
	case commandExists("zenity"):
		cmd = exec.Command("zenity", "--file-selection", "--save", "--confirm-overwrite",
			"--filename="+defaultPath,
			"--title=Save Odoo CSV Report")
	case commandExists("kdialog"):
		cmd = exec.Command("kdialog", "--getsavefilename", defaultPath, "*.csv")
	case commandExists("osascript"):
		script := fmt.Sprintf(`
			set defaultPath to POSIX file "%s"
			set saveFile to choose file name with prompt "Save Odoo CSV Report" default name "%s" default location (path to home folder)
			return POSIX path of saveFile
		`, defaultPath, defaultFileName)
		cmd = exec.Command("osascript", "-e", script)
	default:
		return sourceFile, nil
	}

	output, err := cmd.Output()
	if err != nil {
		return sourceFile, nil
	}

	targetPath := strings.TrimSpace(string(output))
	if targetPath == "" {
		return sourceFile, nil
	}

	if err := copyFile(sourceFile, targetPath); err != nil {
		return "", fmt.Errorf("failed to copy file: %w", err)
	}

	os.Remove(sourceFile)

	return targetPath, nil
}

// OpenPDFSaveDialog opens a save dialog for PDF files
func OpenPDFSaveDialog(sourceFile string) (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	defaultFileName := filepath.Base(sourceFile)
	defaultPath := filepath.Join(homeDir, defaultFileName)

	var cmd *exec.Cmd

	switch {
	case commandExists("zenity"):
		cmd = exec.Command("zenity", "--file-selection", "--save", "--confirm-overwrite",
			"--filename="+defaultPath,
			"--title=Save Mail Report")
	case commandExists("kdialog"):
		cmd = exec.Command("kdialog", "--getsavefilename", defaultPath, "*.pdf")
	case commandExists("osascript"):
		script := fmt.Sprintf(`
			set defaultPath to POSIX file "%s"
			set saveFile to choose file name with prompt "Save Mail Report" default name "%s" default location (path to home folder)
			return POSIX path of saveFile
		`, defaultPath, defaultFileName)
		cmd = exec.Command("osascript", "-e", script)
	default:
		return sourceFile, nil
	}

	output, err := cmd.Output()
	if err != nil {
		return sourceFile, nil
	}

	targetPath := strings.TrimSpace(string(output))
	if targetPath == "" {
		return sourceFile, nil
	}

	return targetPath, nil
}

// commandExists checks if a command is available in PATH
func commandExists(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0644)
}
