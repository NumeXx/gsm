<p align="left">
  <img src="assets/img/gsm-logo.png" alt="GSM Logo" width="40" style="vertical-align: middle; margin-right: 8px;">
  <h3><strong style="font-size: 28px;">GSM - GSocket Manager</strong></h3>
</p>

[![Version](https://img.shields.io/badge/version-v0.3.0-blue)](CHANGELOG.md) [![Go Report Card](https://goreportcard.com/badge/github.com/NumeXx/gsm)](https://goreportcard.com/report/github.com/NumeXx/gsm) [![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

**Connect like there's no firewall, but with style, a kick-ass TUI, and powerful CLI features!**

GSM (GSocket Manager) is a sleek, terminal-based utility to streamline your `gsocket` (Global Socket) connections. Inspired by the simplicity and power of tools like `ssm` (Secure Shell Manager), GSM brings a user-friendly Terminal User Interface (TUI) and a versatile Command Line Interface (CLI) to the world of `gsofcket`, making it a breeze to manage, connect to, and organize your secure, firewall-bypassing endpoints.

Stop fumbling with long `gs-netcat` commands and cryptic keys. With GSM, all your `gsocket` connections are just a few keystrokes or a simple command away.

![GSM TUI Screenshot](assets/img/some.png)

## ✨ Features

*   **Intuitive TUI:** A clean and efficient terminal interface built with BubbleTea.
    *   **Realtime Detail Panel:** View comprehensive details (Key, Tags, Usage, Last Seen) of the selected connection instantly alongside the list.
    *   List display inspired by `ssm` for a dense and informative view (`Name` then `# Tags`).
*   **Full In-TUI Connection Management:**
    *   **Add (`a` key):** Add new GSocket connections. 
        *   **Automatic Mnemonic Names:** If the name is left blank, GSM auto-generates a memorable 3-word name from the GSocket Key!
    *   **Edit (`e` key):** Modify existing connections (name, key, tags) seamlessly.
    *   **Delete (`d` key):** Remove connections with a confirmation step.
*   **Quick Connect:** Select a connection from the TUI list and hit `Enter` – you're in!
*   **Real-time Filtering:** Simply type in the TUI to filter connections by name or tags. Press `Esc` to clear.
*   **Powerful CLI Importer:**
    *   `gsm import --secret "YOUR_KEY#tag1,tag2"`: Import a single GSocket key, with optional tags. A mnemonic name is auto-generated.
    *   `gsm import --file /path/to/keys.txt`: Bulk import GSocket keys from a text file (one `KEY#tag1,tag2` per line). Names are auto-generated, tags are parsed if present.
    *   Intelligent duplicate checking for names and keys during import.
    *   Colorful and informative CLI output.
*   **Connection Tracking:** 
    *   `Usage` count for each connection.
    *   `LastConnected` timestamp automatically updated.
*   **Configuration Storage:** Connections are stored in a human-readable JSON format (`~/.gsm/config.json`).

## 🎯 Who is this for? / Use Cases

GSM is perfect for:

*   **`gsocket` / `gs-netcat` power users:** Manage numerous `gsocket` endpoints without memorizing keys or complex commands.
*   **Pentesters & Security Researchers:** Quickly switch between `gs-netcat` listeners, reverse shells, or other `gsocket`-tunneled services.
*   **Sysadmins & Developers:** Easily access firewalled or NAT-ed machines for remote administration or development via `gsocket`.
*   **Anyone tired of typing long `gs-netcat` commands:** GSM provides a friendly TUI and quick CLI shortcuts.

If you need to reliably connect to services hidden behind firewalls/NATs using the magic of `gsocket`, GSM will be your best friend!

## 🚀 Getting Started

### Prerequisites

*   **Go:** Version 1.18 or higher is recommended.
*   **`gs-netcat`:** (from the [gsocket by The Hackers Choice](https://github.com/hackerschoice/gsocket) suite) must be installed and in your system's `$PATH`.
*   **Supported OS:** Linux, macOS. (Windows, BSDs might work but are not yet officially tested. Contributions welcome!)

### Installation

**Option 1: Using `go install` (Recommended)**

```bash
go install github.com/NumeXx/gsm@latest
```
Make sure your `$GOPATH/bin` or `$HOME/go/bin` is in your system's `$PATH`.

**Option 2: Build from Source**

```bash
git clone https://github.com/NumeXx/gsm.git # Replace with your fork if you prefer
cd gsm
go build -o gsm ./cmd/gsm/
# Optional: Move the binary to your PATH
sudo mv gsm /usr/local/bin/
```

**(Coming Soon: Pre-compiled binaries & install scripts)**

### Quick Usage

1.  **Launch GSM TUI:**
    ```bash
    gsm
    ```
2.  **Import a single secret key (with optional tags):**
    ```bash
    gsm import -s "YOUR_GSOCKET_KEY#project,client-x"
    # Or generate a key and import it immediately (example for bash/zsh)
    gsm import -s "$(gs-netcat -g)#my-temp-server,lab"
    ```
3.  **Import multiple keys from a file:**
    Create a file (e.g., `my_keys.txt`):
    ```
    secretkey1#work,server1
    anothersecretkey2#personal
    justakey3
    ```
    Then run:
    ```bash
    gsm import -f my_keys.txt
    ```

### TUI Keybindings (Main List)

*   **`↑` / `↓` / `j` / `k`**: Navigate connections.
*   **`Enter`**: Connect to the selected endpoint.
*   **`/`**: Enter filter mode (type to filter, `Esc` to clear).
*   **`a`**: Add a new connection.
*   **`e`**: Edit the selected connection.
*   **`d`**: Delete the selected connection (with confirmation).
*   **`q` / `Ctrl+C`**: Quit GSM.

## 🛠️ Configuration

GSM stores its configuration in `~/.gsm/config.json`. While you can view it, using the in-TUI features (`a`, `e`, `d`) or CLI `import` commands is recommended for modifications.

**Example `config.json` entry:**
```json
    {
      "name": "GeneratedMnemonicName",
      "key": "your-actual-gsocket-secret-key",
      "tags": ["imported", "awesome"],
      "usage": 1,
      "last_connected": "2023-10-28T10:30:00Z"
    }
```

## 🛣️ Future Enhancements (To-Do / Ideas)

*   **More Robust `LastConnected` Formatting:** Relative time (e.g., "5 minutes ago").
*   **In-TUI "View Full Key" Option:** For very long keys in the detail panel.
*   **Sort Connections:** By name, last used, usage count within TUI.
*   **Direct `gs-netcat -g` Integration in TUI Add Form:** Option to generate key directly.
*   **Export Connections:** CLI command to export to a plain text key file.
*   **Theme/Color Customization.**
*   **Support for `gsocket` (the wrapper) in addition to `gs-netcat` for connection execution.**

## 🤝 Contributing

Contributions, issues, and feature requests are welcome! Please feel free to check the [issues page](https://github.com/NumeXx/gsm/issues) (or create one!).

1.  Fork the Project
2.  Create your Feature Branch (`git checkout -b feature/AmazingFeature`)
3.  Commit your Changes (`git commit -m 'Add some AmazingFeature'`)
4.  Push to the Branch (`git push origin feature/AmazingFeature`)
5.  Open a Pull Request

## 📜 License

Distributed under the [MIT License](LICENSE). (Ensure you have a `LICENSE` file in your repo).

## 🙏 Acknowledgements

*   **[gsocket](https://github.com/hackerschoice/gsocket) by The Hackers Choice (THC) & Skyper:** For the incredible Global Socket technology that makes this tool possible and for the direct inspiration and feedback.
*   **[ssm (Secure Shell Manager)](https://github.com/lfaoro/ssm) by Leonardo Faoro:** For the brilliant TUI concept and a great example of a connection manager and (etc... hehehe).
*   **[BubbleTea](https://github.com/charmbracelet/bubbletea) & [Lipgloss](https://github.com/charmbracelet/lipgloss) by Charm:** For the fantastic Go libraries that power the TUI.

---
**GSM - Making `gsocket` a joy to manage!**