# Use the official Golang image
FROM golang:1.24

# Install clang
RUN DEBIAN_FRONTEND="noninteractive" apt-get update -y && apt-get install -y \
    clang \
    curl \
    netcat-traditional \
    ssh \
    git \
    net-tools \
    htop \
    build-essential \
    wget

RUN apt-get clean && rm -rf /var/lib/apt/lists/*

# Install GitHub CLI
RUN ARCH=$(dpkg --print-architecture) && \
    GH_VERSION=$(curl -s https://api.github.com/repos/cli/cli/releases/latest | grep tag_name | cut -d '"' -f 4 | cut -c 2-) && \
    curl -sSL "https://github.com/cli/cli/releases/latest/download/gh_${GH_VERSION}_linux_${ARCH}.deb" -o gh.deb && \
    dpkg -i gh.deb && \
    rm gh.deb

# Install Goose CLI (non-interactive)
RUN curl -fsSL https://github.com/block/goose/releases/download/stable/download_cli.sh | CONFIGURE=false bash

# Install Aider
RUN curl -LsSf https://aider.chat/install.sh | sh

# At build time, we don't need to initialize Go module or download dependencies
# because postCreateCommand will handle it.

# The CMD command is used as a default command to run if no other command is specified
# when creating a container. Adjust it as needed for your workflow.
# For example, you might use it to start your application, or in this case,
# to keep the container running and provide instructions.
CMD ["bash", "-c", "echo 'Go container up! Use VS Code terminal to interact.' && tail -f /dev/null"]


