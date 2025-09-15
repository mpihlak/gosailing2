# 🚢 GitHub Pages Deployment Guide

This guide explains how to deploy your Go Sailing Game to GitHub Pages, making it accessible as a static website at `https://yourusername.github.io/your-repo-name`.

## 🚀 Quick Start

### Option 1: Automatic Deployment (Recommended)

1. **Push to GitHub**: Make sure your repository is on GitHub
2. **Enable GitHub Actions**:
   - Go to your repository Settings
   - Click on "Pages" in the left sidebar
   - Under "Source", select "GitHub Actions"
3. **Push changes**: The workflow will automatically trigger on pushes to `main` or `master` branch
4. **Access your game**: Visit `https://yourusername.github.io/your-repo-name`

### Option 2: Manual Build and Deploy

1. **Build the web version**:
   ```bash
   ./build-web.sh
   ```

2. **Commit the web files**:
   ```bash
   git add web/
   git commit -m "Update web build"
   git push
   ```

3. **Enable GitHub Pages**:
   - Go to repository Settings > Pages
   - Set source to "Deploy from a branch"
   - Select `main` branch and `/web` folder
   - Click Save

## 📱 Mobile Support

Your sailing game now includes full mobile touch controls:

- **Steering**: Touch and drag on the left/right sides of the screen
- **Buttons**: Tap the bottom buttons for pause, restart, and help
- **Responsive**: Works on phones, tablets, and desktop browsers

## 🛠️ Files Included

- `web/index.html` - Game webpage with responsive design
- `web/sailing.wasm` - Compiled WebAssembly game binary
- `web/wasm_exec.js` - Go WebAssembly runtime
- `.github/workflows/deploy.yml` - GitHub Actions workflow
- `build-web.sh` - Manual build script

## 🔧 Customization

### Update Game Title
Edit `web/index.html` and modify the `<title>` and header text.

### Change Styling
The CSS is embedded in `index.html`. Modify the `<style>` section to customize:
- Colors and themes
- Layout and responsiveness
- Loading and error messages

### Domain Setup
To use a custom domain:
1. Add a `CNAME` file to the `web/` directory with your domain
2. Configure DNS to point to GitHub Pages
3. Enable HTTPS in repository settings

## 🐛 Troubleshooting

### Build Fails
- Ensure Go 1.21+ is installed
- Check that all dependencies are available: `go mod tidy`
- Verify the build works locally: `./build-web.sh`

### Game Doesn't Load
- Check browser console for errors
- Ensure WASM is supported (modern browsers only)
- Verify all files are accessible (check network tab)

### Mobile Controls Not Working
- Make sure you're using a touch device
- Check that the mobile controls code is integrated
- Test with browser developer tools mobile emulation

## 📊 Performance

The WASM file is approximately 12MB. Consider:
- Using a CDN for faster global access
- Implementing loading progress indicators
- Adding service worker for offline capability

## 🔗 Example URLs

- **Development**: `http://localhost:8080` (using `./serve-web.sh`)
- **GitHub Pages**: `https://username.github.io/ebiten-sailing`
- **Custom Domain**: `https://sailing.yourdomain.com`

## 📝 Next Steps

1. **Test on devices**: Try the game on various phones and tablets
2. **Share and iterate**: Get feedback from players
3. **Monitor usage**: Set up analytics if desired
4. **Add features**: Expand the game based on player feedback

Your sailing game is now ready to sail the web! ⛵🌊
