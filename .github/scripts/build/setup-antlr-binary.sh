#!/bin/bash
set -e

echo '#!/bin/bash' | sudo tee /usr/local/bin/antlr4
echo 'java -jar /usr/local/lib/antlr.jar "$@"' | sudo tee -a /usr/local/bin/antlr4
sudo chmod +x /usr/local/bin/antlr4
