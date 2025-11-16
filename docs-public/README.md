# Documentation Website

This directory contains the web-based documentation viewer for the Simulated Exchange project.

## Structure

```
docs-public/
├── index.html         # Main documentation hub with search and categories
├── view.html          # Markdown viewer that renders .md files as HTML
├── markdown/          # All project markdown files (36 files)
└── README.md          # This file
```

## How It Works

### Main Index (index.html)
- Displays categorized list of all documentation
- Search functionality to filter documents
- Organized into: Main Docs, SRE & Performance, Setup & Configuration
- Links to individual documents via the viewer

### Markdown Viewer (view.html)
- Renders markdown files as GitHub-flavored HTML
- Uses marked.js for client-side rendering
- Automatic table of contents for longer documents
- **Smart Link Rewriting**: Automatically converts markdown links to use the viewer

### Link Rewriting

The viewer automatically rewrites internal markdown links:

**Before (in markdown file):**
```markdown
[Architecture](docs/ARCHITECTURE.md)
[API Reference](API.md)
[Troubleshooting](./TROUBLESHOOTING.md)
```

**After (rendered in browser):**
```
view.html?doc=ARCHITECTURE.md
view.html?doc=API.md
view.html?doc=TROUBLESHOOTING.md
```

This ensures all inter-document links work seamlessly in the web viewer, regardless of the original path structure in the markdown files.

## Features

- ✅ GitHub-flavored markdown rendering
- ✅ Syntax highlighting for code blocks
- ✅ Automatic table of contents
- ✅ Search across all documents
- ✅ Smart link rewriting for inter-document navigation
- ✅ Download raw markdown option
- ✅ Mobile responsive design
- ✅ Dark mode support (via GitHub markdown CSS)

## Usage

### View Documentation Hub
```
https://docs.andythomas-sre.com
```

### View Specific Document
```
https://docs.andythomas-sre.com/view.html?doc=ARCHITECTURE.md
```

### Download Raw Markdown
```
https://docs.andythomas-sre.com/markdown/ARCHITECTURE.md
```

## Adding New Documentation

1. Add your `.md` file to the project (anywhere)
2. Run the copy script to update the web docs:
   ```bash
   find /home/andy/simulated_exchange -name "*.md" -type f \
     ! -path "*/docs-public/*" ! -path "*/.git/*" \
     -exec cp {} /home/andy/simulated_exchange/docs-public/markdown/ \;
   ```
3. (Optional) Update `index.html` to add the document to the categorized list

## Technical Details

- **Markdown Parser**: marked.js v4+ (GitHub-flavored markdown)
- **CSS Framework**: github-markdown-css v5
- **Hosting**: Caddy file server via docker-compose
- **Volume Mount**: Read-only mount to Caddy container at `/usr/share/caddy/docs`

## Configuration

The documentation site is served by Caddy. See `docker/Caddyfile` for the configuration:

```caddyfile
docs.andythomas-sre.com {
    root * /usr/share/caddy/docs
    file_server browse
    try_files {path} {path}/ /index.html
    # ... additional config
}
```

Volume mount in `docker-compose.yml`:
```yaml
volumes:
  - ./docs-public:/usr/share/caddy/docs:ro
```

## Troubleshooting

### Links Don't Work
- The link rewriting happens client-side via JavaScript
- Check browser console for any errors
- Ensure the target markdown file exists in the `markdown/` directory

### Document Not Found
- Verify the file exists: `ls docs-public/markdown/FILENAME.md`
- Re-run the copy script to sync markdown files
- Check that filename in URL matches exactly (case-sensitive)

### Styling Issues
- The site uses CDN-hosted CSS libraries
- Check network tab to ensure CDN resources load
- Verify internet connectivity for external resources

## Maintenance

### Update All Documentation
```bash
# From project root
find . -name "*.md" -type f ! -path "*/docs-public/*" ! -path "*/.git/*" \
  -exec cp {} ./docs-public/markdown/ \;
```

### Test Locally
```bash
# Serve with Python
cd docs-public
python3 -m http.server 8000
# Visit http://localhost:8000
```

### Check Link Rewriting
Open browser console on any document page and you'll see:
```
Rewriting: docs/ARCHITECTURE.md -> view.html?doc=ARCHITECTURE.md
```

## Browser Compatibility

- Chrome/Edge: ✅ Fully supported
- Firefox: ✅ Fully supported
- Safari: ✅ Fully supported
- Mobile browsers: ✅ Responsive design

Minimum versions:
- Chrome 80+
- Firefox 75+
- Safari 13+
- Edge 80+
