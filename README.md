# GSM - GSocket Manager

**Connect like there's no firewall, but with style and a kick-ass TUI!**

GSM (GSocket Manager) is a sleek, terminal-based ankkutilitas/manajer to streamline your `gsocket` (Global Socket) connections. Inspired by the simplicity of tools like `ssm` (Secure Shell Manager), GSM brings a user-friendly Terminal User Interface (TUI) to the world of `gsocket`, making it a breeze to manage, connect to, and organize your secure, firewall-bypassing endpoints.

Stop fumbling with long `gs-netcat` commands and cryptic keys. With GSM, all your `gsocket` connections are just a few keystrokes away.

## ✨ Features

*   **Intuitive TUI:** A clean and efficient terminal interface built with BubbleTea to manage your connections.
*   **Connection Management:** Easily add, (and soon edit/delete) your `gsocket` connection configurations.
*   **Quick Connect:** Select a connection from the list and hit Enter – you're in!
*   **Configuration Storage:** Connections are stored मानव-readable JSON format (`~/.gsm/config.json`).
*   **Tagging Support:** Organize your connections with tags for better filtering and identification (filtering coming soon!).
*   **(Planned) Real-time Filtering:** Quickly find connections by name or tag.
*   **(Planned) In-TUI Configuration Editing:** Modify your connections without leaving GSM.

## 🚀 Getting Started

### Prerequisites

*   Go (version 1.18+ recommended)
*   `gs-netcat` (from the [gsocket](https://github.com/hackerschoice/gsocket) suite) must be installed and in your system's PATH.

### Installation

1.  **Clone the repository:**
    ```bash
    git clone https://github.com/NumeXx/gsm.git # Make sure to use your actual repo URL!
    cd gsm
    ```

2.  **Build the binary:**
    ```bash
    go build -o gsm ./cmd/gsm/
    ```
    This will create a `gsm` binary in the current directory. You can move this to a directory in your `$PATH` (e.g., `/usr/local/bin` or `~/bin`) to make it accessible globally.
    ```bash
    # Example:
    # sudo mv gsm /usr/local/bin/
    ```

### Quick Usage

1.  **Add your first GSocket connection:**
    ```bash
    gsm config
    ```
    Follow the prompts to enter:
    *   **Connection name:** A friendly name for your connection (e.g., `my-remote-server`).
    *   **GSocket key (-s value):** The secret key for your `gsocket` endpoint.
    *   **Tags (comma separated):** Optional tags to help you organize (e.g., `work,bastion,dev`).

2.  **Launch GSM TUI:**
    ```bash
    gsm
    ```
    Navigate the list using arrow keys (↑/↓), and press `Enter` to connect to the selected GSocket endpoint. Press `q` or `Ctrl+C` to quit.

## 🛠️ Configuration

GSM stores its configuration in `~/.gsm/config.json`. It's a simple JSON file you can also edit manually if needed (though using `gsm config` is recommended).

**Example `config.json`:**
```json
{
  "connections": [
    {
      "name": "my-server-1",
      "key": "your-gsocket-secret-key-1",
      "tags": ["work", "production"],
      "usage": 0
    },
    {
      "name": "home-lab",
      "key": "another-secret-key",
      "tags": ["personal", "lab"],
      "usage": 0
    }
  ]
}
```

## 🤝 Contributing

Contributions, issues, and feature requests are welcome! Feel free to check the [issues page](https://github.com/NumeXx/gsm/issues) (if you have one).

1.  Fork the Project
2.  Create your Feature Branch (`git checkout -b feature/AmazingFeature`)
3.  Commit your Changes (`git commit -m 'Add some AmazingFeature'`)
4.  Push to the Branch (`git push origin feature/AmazingFeature`)
5.  Open a Pull Request

## 📜 License

Distributed under the [MIT License](LICENSE). See `LICENSE` file for more information. (You'll need to add a LICENSE file, MIT is a good default for open source).

## 🙏 Acknowledgements

*   [gsocket](https://github.com/hackerschoice/gsocket) by The Hackers Choice.
*   [BubbleTea](https://github.com/charmbracelet/bubbletea) & [Lipgloss](https://github.com/charmbracelet/lipgloss) by Charm.
*   Inspiration from [ssm](https://github.com/lfaoro/ssm).

---

**GSM - Get in, get out, no fuss. Just pure, unadulterated `gsocket` connectivity.**
