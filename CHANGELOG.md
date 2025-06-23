# 📝 Changelog Generation

This project includes a convenient local script to generate detailed changelogs for releases!

## 🚀 Quick Start

### Generate Changelog Locally
Use the local script to generate and preview your changelog before releasing:

```bash
# Generate changelog from last release to HEAD
./scripts/generate-changelog.sh

# Generate changelog between specific tags
./scripts/generate-changelog.sh v0.0.7 HEAD

# Generate changelog between specific commits
./scripts/generate-changelog.sh abc123 def456
```

### Use in GitHub Release
1. Run the local script to generate your changelog
2. Copy the output
3. Go to **Actions** → **Build and Push Docker Images Manually**
4. Enter your version (e.g., `1.2.3`)
5. Paste the generated changelog into the changelog field
6. Run the workflow

## 📋 Changelog Format

The generated changelogs include:

- **Categorized commits** with emojis (🚀 features, 🐛 fixes, etc.)
- **PR numbers or commit hashes** for easy reference
- **Contributor attribution** with GitHub usernames
- **Commit count** and **contributor count** since last release
- **Detailed categorization** by change type

## 💡 Tips for Better Changelogs

### Use Conventional Commits
Format your commit messages with prefixes for better categorization:

```
feat: add new monitoring dashboard
fix: resolve memory leak in heartbeat service
docs: update API documentation
style: improve button styling
refactor: optimize database queries
test: add unit tests for auth service
chore: update dependencies
```

### Categories & Emojis
- `feat:` / `feature:` → 🚀 **New Features**
- `fix:` / `bug:` → 🐛 **Bug Fixes**
- `docs:` / `doc:` → 📚 **Documentation**
- `style:` / `ui:` / `design:` → 💄 **UI/Style**
- `refactor:` / `perf:` → ⚡ **Improvements**
- `test:` / `tests:` → 🧪 **Tests**
- `chore:` / `build:` / `ci:` → 🔧 **Maintenance**

## 🔧 How It Works

1. **Auto-detection**: Finds the latest release tag automatically
2. **Git log parsing**: Extracts commits between releases
3. **Categorization**: Groups commits by type (feat, fix, etc.)
4. **Contributor detection**: Extracts GitHub usernames from commit emails
5. **PR detection**: Identifies merge commits and extracts PR numbers
6. **Formatting**: Adds emojis and clean formatting

## 📚 Examples

### Example Generated Changelog
```
## 🚀 New Features
#123 feat: add Ntfy notification channel integration (Thanks @0xfurai)

## 🐛 Bug Fixes
#124 fix: remove default value for priority in Ntfy form schema (Thanks @0xfurai)

## 📚 Documentation
#125 docs: update README to include beta status (Thanks @0xfurai)
#126 docs: enhance README with additional badges (Thanks @0xfurai)

## 🔧 Maintenance
#127 chore: update port mapping in docker-compose.prod.yml (Thanks @0xfurai)

## 📊 Release Statistics
- **13** commits since v0.0.8
- **3** contributors

## 👥 Contributors
Thanks to: @0xfurai @dbrennand @Yevhen Piotrovskyi
```

### Example Local Script Output
```bash
$ ./scripts/generate-changelog.sh

🎉 Peekaping Detailed Changelog Generator
==========================================

📋 Generating detailed changelog from v0.0.8 to HEAD

## 🚀 New Features
0b3a16f feat: add Ntfy notification channel integration (Thanks @0xfurai)

## 🐛 Bug Fixes
dbbdb62 fix: remove default value for priority in Ntfy form schema (Thanks @0xfurai)

## 📚 Documentation
e12ad76 docs: update README to include beta status (Thanks @0xfurai)
3cf122d docs: enhance README with additional badges (Thanks @0xfurai)

## 🔧 Maintenance
517b484 chore: update port mapping in docker-compose.prod.yml (Thanks @0xfurai)

## 📊 Release Statistics
- **13** commits since v0.0.8
- **3** contributors

## 👥 Contributors
Thanks to: @0xfurai @dbrennand @Yevhen Piotrovskyi

==========================================
✅ Detailed changelog generated successfully!

💡 Usage tips:
• Copy the sections above for your GitHub release
• Use conventional commit messages (feat:, fix:, docs:, etc.) for better categorization
• PR numbers will be automatically detected from merge commits

🚀 Ready to release? Run the GitHub Actions workflow with version number!
```

## 🛠 Customization

You can modify the changelog generation by editing:
- `scripts/generate-changelog.sh` - Local script

## 📝 Release Workflow

1. **Generate changelog**: Run `./scripts/generate-changelog.sh`
2. **Review output**: Check the categorized changes
3. **Copy changelog**: Select and copy the relevant sections
4. **Create release**: Use GitHub Actions with your changelog
5. **Publish**: The workflow will build and push Docker images

Happy releasing! 🎉
