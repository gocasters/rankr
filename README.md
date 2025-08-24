# rankr
Rank repo contributors


---
## Development Setup
Development Environment Setup

This project uses ASDF to manage tool versions. Follow these steps to set up your environment:

### 1. Install ASDF

Run these commands in your terminal:

```bash 
# Install ASDF
git clone https://github.com/asdf-vm/asdf.git ~/.asdf --branch v0.14.0

# Add to shell (choose one based on your shell)
echo -e '\n. "$HOME/.asdf/asdf.sh"' >> ~/.bashrc  # Bash
echo -e '\n. "$HOME/.asdf/asdf.sh"' >> ~/.zshrc   # Zsh
source ~/.bashrc  # or restart terminal

# Verify installation
asdf --version  # Should print "v0.14.0" or higher
```

### 2. Install Plugins & Tools

This project requires:
Run:
```bash 
# Add plugins
asdf plugin add golang https://github.com/asdf-community/asdf-golang.git

# Install versions from .tool-versions
asdf install
```
Verify installations:
```bash
go version
```

### 3. Project Setup

After cloning the repository:

```bash
git clone https://github.com/your-org/rankr.git
cd rankr
asdf install  # Install tools
```
---