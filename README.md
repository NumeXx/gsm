# GSM - GSocket Manager

[![Version](https://img.shields.io/badge/version-v0.2.0-blue)](CHANGELOG.md)

**Connect like there's no firewall, but with style and a kick-ass TUI!**

GSM (GSocket Manager) is a sleek, terminal-based utility to streamline your `gsocket` (Global Socket) connections. Inspired by the simplicity of tools like `ssm` (Secure Shell Manager), GSM brings a user-friendly Terminal User Interface (TUI) to the world of `gsocket`, making it a breeze to manage, connect to, and organize your secure, firewall-bypassing endpoints.

Stop fumbling with long `gs-netcat` commands and cryptic keys. With GSM, all your `gsocket` connections are just a few keystrokes away.

## ✨ Features

*   **Intuitive TUI:** A clean and efficient terminal interface built with BubbleTea to manage your connections.
*   **Full In-TUI Connection Management:**
    *   **Add (`a` key):** Add new GSocket connections directly within the TUI.
    *   **Edit (`e` key):** Modify existing connections (name, key, tags) seamlessly.
    *   **Delete (`d` key):** Remove connections with a confirmation step.
*   **Quick Connect:** Select a connection from the list and hit `Enter` – you're in!
*   **Real-time Filtering:** Simply type to filter connections by name or tags. Press `Esc` to clear the filter.
*   **Configuration Storage:** Connections are stored in a human-readable JSON format (`~/.gsm/config.json`).
*   **Tagging Support:** Organize your connections with tags.

## 🚀 Getting Started

### Prerequisites

*   Go (Golang)
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

### Quick Usage & TUI Keybindings

1.  **Launch GSM TUI:**
    ```bash
    gsm
    ```
    The TUI will launch, displaying your GSocket connections.

2.  **Navigating the List:**
    *   **`↑` / `↓`**: Navigate up and down the connection list.
    *   **`Enter`**: Connect to the selected GSocket endpoint.
    *   **`q` / `Ctrl+C`**: Quit GSM.

3.  **Filtering Connections:**
    *   **Type `/` to Filter**: Simply start typing any part of the connection's name or tags. The list will filter obstáculos real-time.
    *   **`Esc` (while filtering)**: Clear the current filter and show all connections.

4.  **Managing Connections (from the list view):**
    *   **`a`**: Add a new GSocket connection. This will open a form.
        *   In the form: Use `Tab` / `Shift+Tab` to navigate fields. `Enter` to save, `Esc` to cancel.
    *   **`e`**: Edit the selected GSocket connection. This will open a form pre-filled with the connection's details.
        *   In the form: Use `Tab` / `Shift+Tab` to navigate fields. `Enter` to save, `Esc` to cancel.
    *   **`d`**: Delete the selected GSocket connection. A confirmation prompt (`y/N`) will appear.

## 🛠️ Configuration

GSM stores its configuration in `~/.gsm/config.json`. It's a simple JSON file. While you can edit it manually, using the in-TUI features (`a` to add, `e` to edit) is now the recommended way.

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

---
**GSM - Get in, get out, no fuss. Just pure, unadulterated `gsocket` connectivity.**

### 🛣️ Future Enhancements (To-Do)

*   **Quick Connect to Last Used:** Shortcut or command to immediately connect to the last used GSocket connection.
*   **Quick Connect to Most Used:** Shortcut or command to connect to the most frequently used connection (based on `usage` count).
*   **Advanced Search/Sort:**
    *   Sort connections by name, last used, or usage count.
    *   More specific search queries (e.g., `tag:work status:active`).
*   **In-TUI 'View Details':** Display connection details without entering edit mode (e.g., with `ctrl+v`).
*   **UI Polish:**
    *   Optional borders around the list.
    *   Auto-clearing status messages (success/error).
    *   More theme/color options.

## 🤝 Contributing

Contributions, issues, and feature requests are welcome! Feel free to check the [issues page](https://github.com/NumeXx/gsm/issues).

1.  Fork the Project
2.  Create your Feature Branch (`git checkout -b feature/AmazingFeature`)
3.  Commit your Changes (`git commit -m 'Add some AmazingFeature'`)
4.  Push to the Branch (`git push origin feature/AmazingFeature`)
5.  Open a Pull Request

## 📜 License

Distributed under the [MIT License](LICENSE). See `LICENSE` file for more information. (Remember to add a `LICENSE` file).

## 🙏 Acknowledgements

*   [gsocket](https://github.com/hackerschoice/gsocket) by The Hackers Choice.
*   [BubbleTea](https://github.com/charmbracelet/bubbletea) & [Lipgloss](https://github.com/charmbracelet/lipgloss) by Charm.
*   Inspiration from [ssm](https://github.com/lfaoro/ssm).