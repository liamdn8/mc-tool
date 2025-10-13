#!/bin/bash

echo "╔════════════════════════════════════════════════════════════════╗"
echo "║        MC-Tool Web UI - Interactive Demo                      ║"
echo "╚════════════════════════════════════════════════════════════════╝"
echo

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${BLUE}🌐 MC-Tool Web UI Features:${NC}"
echo
echo "  ✅ Bilingual Interface (English & Vietnamese)"
echo "  ✅ Dashboard with MinIO aliases overview"
echo "  ✅ Visual bucket comparison tool"
echo "  ✅ Bucket analysis with statistics"
echo "  ✅ Memory profiling and leak detection"
echo "  ✅ Configuration checklist"
echo "  ✅ Real-time job tracking"
echo

echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo

echo -e "${BLUE}📋 Available Commands:${NC}"
echo
echo "  # Start web UI on default port (8080)"
echo "  $ mc-tool web"
echo
echo "  # Start on custom port"
echo "  $ mc-tool web --port 3000"
echo
echo "  # Get help"
echo "  $ mc-tool web --help"
echo

echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo

echo -e "${BLUE}🚀 Quick Start:${NC}"
echo
echo "  1. Start the web server:"
echo "     $ ./build/mc-tool-portable web"
echo
echo "  2. Open your browser:"
echo "     http://localhost:8080"
echo
echo "  3. Select your language:"
echo "     🇬🇧 English or 🇻🇳 Tiếng Việt"
echo
echo "  4. Use the sidebar to navigate:"
echo "     • Dashboard   - Overview and status"
echo "     • Compare     - Compare buckets"
echo "     • Analyze     - Analyze bucket contents"
echo "     • Profile     - Memory and performance profiling"
echo "     • Checklist   - Configuration validation"
echo

echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo

echo -e "${BLUE}📚 Example Operations:${NC}"
echo

echo -e "${GREEN}Compare Buckets:${NC}"
echo "  Source:      minio1/bucket1/folder"
echo "  Destination: minio2/bucket2/folder"
echo "  ✓ Enable recursive comparison"
echo "  → Click 'Start Comparison'"
echo

echo -e "${GREEN}Analyze Bucket:${NC}"
echo "  1. Select alias from dropdown"
echo "  2. Select bucket (auto-loads)"
echo "  3. Optional: enter prefix"
echo "  → Click 'Start Analysis'"
echo

echo -e "${GREEN}Profile Server:${NC}"
echo "  Type:     Heap (for memory leak detection)"
echo "  Duration: 5m"
echo "  ✓ Enable leak detection"
echo "  Interval: 10s"
echo "  Threshold: 50 MB"
echo "  → Click 'Start Profiling'"
echo

echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo

echo -e "${BLUE}🔧 Technical Details:${NC}"
echo
echo "  Binary Size:   ~11 MB (portable, statically linked)"
echo "  Dependencies:  None (single binary)"
echo "  Backend:       Go HTTP server"
echo "  Frontend:      HTML5 + CSS3 + Vanilla JS"
echo "  API:           RESTful JSON endpoints"
echo "  i18n:          Built-in English & Vietnamese"
echo

echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo

echo -e "${GREEN}✨ Ready to start?${NC}"
echo
read -p "Press Enter to start the web UI (or Ctrl+C to exit)..."
echo

echo -e "${BLUE}🚀 Starting MC-Tool Web UI...${NC}"
echo

./build/mc-tool-portable web
