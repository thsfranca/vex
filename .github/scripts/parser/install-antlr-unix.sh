#!/bin/bash
set -e

# Install ANTLR for Linux/macOS
wget https://www.antlr.org/download/antlr-${ANTLR_VERSION}-complete.jar -O /tmp/antlr.jar
sudo mkdir -p /usr/local/lib
sudo mv /tmp/antlr.jar /usr/local/lib/
echo '#!/bin/bash' | sudo tee /usr/local/bin/antlr4
echo 'java -jar /usr/local/lib/antlr.jar "$@"' | sudo tee -a /usr/local/bin/antlr4
sudo chmod +x /usr/local/bin/antlr4