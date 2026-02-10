# LazyMan

A fast, beautiful TUI (Terminal User Interface) for browsing and reading manual pages, inspired by [lazygit](https://github.com/jesseduffield/lazygit).

![License](https://img.shields.io/badge/license-MIT-blue.svg)
![Go Version](https://img.shields.io/badge/go-%3E%3D1.25-blue.svg)

## Features

- 🚀 **Fast Navigation** - Quickly browse through all available man pages
- 🔍 **Instant Search** - Search man pages by keyword
- 📖 **Smooth Reading** - Read man pages with vim-style navigation
- 🎨 **Beautiful UI** - Color-coded interface built with Bubble Tea
- ⌨️ **Keyboard-driven** - Efficient keyboard shortcuts for power users
- 🗂️ **Deep Search** (In Beta) - A deep search using pre-built indices for in-depth lookups.

## Installation

```bash
git clone https://github.com/shuklasaharsh/lazyman
cd lazymanuals
go build
```

Or install directly:

```bash
go install github.com/shuklasaharsh/lazyman@latest
```

## Usage

Simply run:

```bash
./lazyman
```

Or if installed:

```bash
lazyman
```

Deep Search:
- For the First time:
```bash
lazyman -S # To build indices
```

- Next time
```bash
lazyman -S uv_loop 
```

### Keyboard Shortcuts

#### List View
- `↑/k` - Move up
- `↓/j` - Move down
- `Enter` - View selected man page
- `/` - Search man pages
- `r` - Refresh man page list
- `q` - Quit

#### Detail View
- `↑/k` - Scroll up
- `↓/j` - Scroll down
- `g` - Go to top
- `G` - Go to bottom
- `u` - Half page up
- `d` - Half page down
- `q/Esc` - Back to list

#### Search View
- `Enter` - Execute search
- `Esc` - Cancel search

## Requirements

- Go 1.25 or higher
- Unix-like system with `man` command (Linux, macOS, BSD)

## Dependencies

- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [Lipgloss](https://github.com/charmbracelet/lipgloss) - Style definitions
- [Bubbles](https://github.com/charmbracelet/bubbles) - TUI components
- [Bleve](https://github.com/blevesearch/bleve) - Full-text search engine

## Contributing

We welcome contributions! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for details on how to get started.

## License

BSD License - see LICENSE file for details

## Acknowledgments

- Inspired by [lazygit](https://github.com/jesseduffield/lazygit)
- Built with [Charm](https://github.com/charmbracelet) tools

## Support

If you find a bug or have a feature request, please open an issue on GitHub.
