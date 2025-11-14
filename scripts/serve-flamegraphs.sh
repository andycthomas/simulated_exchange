#!/bin/bash
################################################################################
# Flamegraph Web Server
#
# Starts a simple HTTP server to view flamegraphs remotely
#
# Usage: ./serve-flamegraphs.sh [port]
#   port: HTTP server port (default: 8888)
################################################################################

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
FLAMEGRAPH_DIR="$PROJECT_ROOT/flamegraphs"

PORT="${1:-8888}"

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'

echo "╔═══════════════════════════════════════════════════════════════════════════╗"
echo "║                    FLAMEGRAPH WEB SERVER                                  ║"
echo "╚═══════════════════════════════════════════════════════════════════════════╝"
echo ""

# Check if flamegraph directory exists
if [[ ! -d "$FLAMEGRAPH_DIR" ]]; then
    echo -e "${YELLOW}⚠️  Flamegraph directory not found: $FLAMEGRAPH_DIR${NC}"
    echo ""
    read -p "Create it now? (Y/n): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Nn]$ ]]; then
        mkdir -p "$FLAMEGRAPH_DIR"
        echo -e "${GREEN}✓${NC} Created $FLAMEGRAPH_DIR"
    else
        echo "Aborted."
        exit 1
    fi
fi

# Check for flamegraphs
SVG_COUNT=$(find "$FLAMEGRAPH_DIR" -name "*.svg" 2>/dev/null | wc -l)

echo -e "${CYAN}Flamegraph Directory:${NC} $FLAMEGRAPH_DIR"
echo -e "${CYAN}Port:${NC} $PORT"
echo -e "${CYAN}SVG Files Found:${NC} $SVG_COUNT"
echo ""

if [[ $SVG_COUNT -eq 0 ]]; then
    echo -e "${YELLOW}⚠️  No SVG flamegraphs found!${NC}"
    echo ""
    echo "Generate flamegraphs with:"
    echo "  make profile-cpu"
    echo "  make profile-all"
    echo "  make chaos-flamegraph"
    echo ""
fi

# Generate index.html
INDEX_FILE="$FLAMEGRAPH_DIR/index.html"

cat > "$INDEX_FILE" <<'EOF'
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Flamegraph Viewer</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: #333;
            padding: 20px;
        }
        .container {
            max-width: 1400px;
            margin: 0 auto;
            background: white;
            border-radius: 12px;
            box-shadow: 0 10px 40px rgba(0,0,0,0.2);
            overflow: hidden;
        }
        header {
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            padding: 30px;
            text-align: center;
        }
        h1 {
            font-size: 2.5em;
            margin-bottom: 10px;
        }
        .subtitle {
            opacity: 0.9;
            font-size: 1.1em;
        }
        .stats {
            display: flex;
            justify-content: space-around;
            padding: 20px;
            background: #f8f9fa;
            border-bottom: 1px solid #dee2e6;
        }
        .stat {
            text-align: center;
        }
        .stat-value {
            font-size: 2em;
            font-weight: bold;
            color: #667eea;
        }
        .stat-label {
            color: #6c757d;
            font-size: 0.9em;
            margin-top: 5px;
        }
        .controls {
            padding: 20px;
            background: #f8f9fa;
            border-bottom: 1px solid #dee2e6;
        }
        .filter-group {
            display: flex;
            gap: 10px;
            align-items: center;
            flex-wrap: wrap;
        }
        .filter-btn {
            padding: 8px 16px;
            border: 2px solid #667eea;
            background: white;
            color: #667eea;
            border-radius: 20px;
            cursor: pointer;
            font-size: 0.9em;
            transition: all 0.3s;
        }
        .filter-btn:hover {
            background: #667eea;
            color: white;
        }
        .filter-btn.active {
            background: #667eea;
            color: white;
        }
        .content {
            padding: 20px;
        }
        .flamegraph-grid {
            display: grid;
            grid-template-columns: repeat(auto-fill, minmax(500px, 1fr));
            gap: 20px;
        }
        .flamegraph-card {
            border: 1px solid #dee2e6;
            border-radius: 8px;
            overflow: hidden;
            transition: all 0.3s;
            background: white;
        }
        .flamegraph-card:hover {
            box-shadow: 0 4px 12px rgba(0,0,0,0.15);
            transform: translateY(-2px);
        }
        .card-header {
            background: #f8f9fa;
            padding: 15px;
            border-bottom: 1px solid #dee2e6;
        }
        .card-title {
            font-weight: bold;
            font-size: 1.1em;
            color: #333;
            margin-bottom: 5px;
        }
        .card-meta {
            font-size: 0.85em;
            color: #6c757d;
        }
        .card-body {
            padding: 0;
        }
        .flamegraph-preview {
            width: 100%;
            height: 300px;
            cursor: pointer;
            border: none;
            display: block;
        }
        .card-footer {
            padding: 15px;
            background: #f8f9fa;
            border-top: 1px solid #dee2e6;
            display: flex;
            gap: 10px;
        }
        .btn {
            padding: 8px 16px;
            border: none;
            border-radius: 4px;
            cursor: pointer;
            font-size: 0.9em;
            transition: all 0.3s;
            text-decoration: none;
            display: inline-block;
        }
        .btn-primary {
            background: #667eea;
            color: white;
        }
        .btn-primary:hover {
            background: #5568d3;
        }
        .btn-secondary {
            background: #6c757d;
            color: white;
        }
        .btn-secondary:hover {
            background: #5a6268;
        }
        .empty-state {
            text-align: center;
            padding: 60px 20px;
            color: #6c757d;
        }
        .empty-state h2 {
            margin-bottom: 15px;
            color: #495057;
        }
        footer {
            text-align: center;
            padding: 20px;
            background: #f8f9fa;
            color: #6c757d;
            border-top: 1px solid #dee2e6;
        }
    </style>
</head>
<body>
    <div class="container">
        <header>
            <h1>🔥 Flamegraph Viewer</h1>
            <div class="subtitle">Interactive Performance Profiling</div>
        </header>

        <div class="stats" id="stats">
            <div class="stat">
                <div class="stat-value" id="total-count">0</div>
                <div class="stat-label">Total Flamegraphs</div>
            </div>
            <div class="stat">
                <div class="stat-value" id="cpu-count">0</div>
                <div class="stat-label">CPU Profiles</div>
            </div>
            <div class="stat">
                <div class="stat-value" id="heap-count">0</div>
                <div class="stat-label">Heap Profiles</div>
            </div>
        </div>

        <div class="controls">
            <div class="filter-group">
                <span style="font-weight: bold;">Filter:</span>
                <button class="filter-btn active" data-filter="all">All</button>
                <button class="filter-btn" data-filter="cpu">CPU</button>
                <button class="filter-btn" data-filter="heap">Heap</button>
                <button class="filter-btn" data-filter="goroutine">Goroutine</button>
                <button class="filter-btn" data-filter="allocs">Allocations</button>
                <button class="filter-btn" data-filter="block">Block</button>
                <button class="filter-btn" data-filter="mutex">Mutex</button>
            </div>
        </div>

        <div class="content">
            <div class="flamegraph-grid" id="flamegraph-grid"></div>
        </div>

        <footer>
            <p>Generated by Simulated Exchange Profiling System</p>
            <p style="margin-top: 5px; font-size: 0.9em;">
                Refresh this page to see new flamegraphs
            </p>
        </footer>
    </div>

    <script>
        // Load flamegraphs
        async function loadFlamegraphs() {
            try {
                const response = await fetch('/');
                const text = await response.text();
                const parser = new DOMParser();
                const doc = parser.parseFromString(text, 'text/html');
                const links = Array.from(doc.querySelectorAll('a'));

                const flamegraphs = links
                    .filter(a => a.href.endsWith('.svg'))
                    .map(a => {
                        const filename = a.textContent.trim();
                        const type = filename.split('_')[0];
                        const timestamp = filename.match(/\d{8}-\d{6}/)?.[0] || 'unknown';
                        return { filename, type, timestamp };
                    });

                renderFlamegraphs(flamegraphs);
                updateStats(flamegraphs);
            } catch (error) {
                console.error('Failed to load flamegraphs:', error);
                document.getElementById('flamegraph-grid').innerHTML = `
                    <div class="empty-state">
                        <h2>No flamegraphs found</h2>
                        <p>Generate flamegraphs using:</p>
                        <code>make profile-cpu</code> or <code>make profile-all</code>
                    </div>
                `;
            }
        }

        function renderFlamegraphs(flamegraphs, filter = 'all') {
            const grid = document.getElementById('flamegraph-grid');

            const filtered = filter === 'all'
                ? flamegraphs
                : flamegraphs.filter(f => f.type === filter);

            if (filtered.length === 0) {
                grid.innerHTML = `
                    <div class="empty-state">
                        <h2>No flamegraphs found</h2>
                        <p>No ${filter} profiles available</p>
                    </div>
                `;
                return;
            }

            grid.innerHTML = filtered.map(fg => `
                <div class="flamegraph-card" data-type="${fg.type}">
                    <div class="card-header">
                        <div class="card-title">${fg.type.toUpperCase()} Profile</div>
                        <div class="card-meta">${fg.timestamp}</div>
                    </div>
                    <div class="card-body">
                        <object class="flamegraph-preview" data="${fg.filename}" type="image/svg+xml"
                                onclick="window.open('${fg.filename}', '_blank')">
                        </object>
                    </div>
                    <div class="card-footer">
                        <a href="${fg.filename}" target="_blank" class="btn btn-primary">Open Full</a>
                        <a href="${fg.filename}" download class="btn btn-secondary">Download</a>
                    </div>
                </div>
            `).join('');
        }

        function updateStats(flamegraphs) {
            document.getElementById('total-count').textContent = flamegraphs.length;
            document.getElementById('cpu-count').textContent =
                flamegraphs.filter(f => f.type === 'cpu').length;
            document.getElementById('heap-count').textContent =
                flamegraphs.filter(f => f.type === 'heap').length;
        }

        // Filter functionality
        document.querySelectorAll('.filter-btn').forEach(btn => {
            btn.addEventListener('click', async () => {
                document.querySelectorAll('.filter-btn').forEach(b => b.classList.remove('active'));
                btn.classList.add('active');

                const filter = btn.dataset.filter;
                const response = await fetch('/');
                const text = await response.text();
                const parser = new DOMParser();
                const doc = parser.parseFromString(text, 'text/html');
                const links = Array.from(doc.querySelectorAll('a'));

                const flamegraphs = links
                    .filter(a => a.href.endsWith('.svg'))
                    .map(a => {
                        const filename = a.textContent.trim();
                        const type = filename.split('_')[0];
                        const timestamp = filename.match(/\d{8}-\d{6}/)?.[0] || 'unknown';
                        return { filename, type, timestamp };
                    });

                renderFlamegraphs(flamegraphs, filter);
            });
        });

        // Load on page load
        loadFlamegraphs();
    </script>
</body>
</html>
EOF

echo -e "${GREEN}✓${NC} Generated index.html"
echo ""

# Determine which Python HTTP server to use
if command -v python3 &> /dev/null; then
    PYTHON_CMD="python3"
elif command -v python &> /dev/null; then
    PYTHON_CMD="python"
else
    echo -e "${YELLOW}⚠️  Python not found, cannot start HTTP server${NC}"
    exit 1
fi

# Get local IP for remote access
LOCAL_IP=$(hostname -I 2>/dev/null | awk '{print $1}' || echo "localhost")
HOSTNAME=$(hostname -f 2>/dev/null || hostname)

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo -e "${GREEN}✓ Starting HTTP server on port $PORT${NC}"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo -e "${CYAN}Access flamegraphs at:${NC}"
echo -e "  Local:      ${BLUE}http://localhost:$PORT/${NC}"
echo -e "  Network:    ${BLUE}http://$LOCAL_IP:$PORT/${NC}"
echo -e "  Hostname:   ${BLUE}http://$HOSTNAME:$PORT/${NC}"
echo ""
echo -e "${CYAN}Directory:${NC} $FLAMEGRAPH_DIR"
echo ""
echo "Press Ctrl+C to stop the server"
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

cd "$FLAMEGRAPH_DIR"
$PYTHON_CMD -m http.server "$PORT"
