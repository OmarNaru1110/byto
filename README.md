# byto

> A modern, lightweight, and powerful GUI for [yt-dlp](https://github.com/yt-dlp/yt-dlp), designed to simplify media downloads.

![byto wallpaper](/assets/byto_ui.png)

<p align="center">
  <img src="https://img.shields.io/github/downloads/OmarNaru1110/byto/total?style=for-the-badge" alt="GitHub Downloads (all releases)">
</p>

**byto** wraps the complexity of the command-line interface into a beautiful, easy-to-use desktop application. Whether you're downloading a single media or archiving a playlist, byto handles it with efficiency and style.

---

## Key Features

- **Modern UI**: Built with React and modern design principles, offering a clean, dark-themed interface.
- **Smart Queue System**: Add multiple URLs, prioritize them, and manage your download queue effortlessly.
- **Parallel Downloads**: Maximize your bandwidth by downloading multiple videos simultaneously.
- **Auto-Dependency Management**: Byto automatically detects, downloads, and updates [yt-dlp](https://github.com/yt-dlp/yt-dlp) and [ffmpeg](https://www.ffmpeg.org/) for you. No manual setup required.
- **Quality Control**: Select your preferred video resolution, from efficient **360p** up to crisp **4K (2160p)**.
- **Real-time Logs**: View detailed logs for every download to understand exactly what's happening.
- **Audio Only**: Extract and download only the audio track from videos.
- **Media Range**: Download a specific time range of a video or audio file.
- **Specific Videos**: Select and download specific videos from a playlist.
- **Playlist Range**: Download a specific range of videos from any playlist.
- **Cookie Support**: Support for cookies to enable authorized downloads.
- **Subtitles Download**: Automatically extract and download subtitles for media.
## Technology Stack

Byto relies on a robust stack to deliver high performance and a small footprint:

- **Backend**: [Go](https://go.dev/) (powered by the [Wails](https://wails.io/) framework)
- **Frontend**: [React](https://react.dev/), [TypeScript](https://www.typescriptlang.org/), and [Vite](https://vitejs.dev/)
- **Styling**: [Tailwind CSS](https://tailwindcss.com/)
- **Core Engine**: [yt-dlp](https://github.com/yt-dlp/yt-dlp)

## Getting Started

### Installation

Download the latest release from the [Releases](https://github.com/OmarNaru1110/byto/releases) page.

#### Windows
1. Download `byto-amd64-installer.exe`
2. You may see a SmartScreen warning (the app is not code-signed). Click **"More info"** → **"Run anyway"**
3. Follow the installation wizard

#### Linux
1. Download `byto-linux-amd64` from the release page
2. Make it executable and run it:

```bash
chmod +x byto-linux-amd64
./byto-linux-amd64
```

Byto is built with Wails, so some Linux systems also need GTK3 and WebKit2GTK runtime libraries installed. If those libraries are already present on your distro, no extra setup is needed. If Byto does not start, run `wails doctor` to see the exact packages your system needs.

Common package names are:

- Debian / Ubuntu 22.04+: `libgtk-3-0 libwebkit2gtk-4.1-0`
- Debian 11 / Ubuntu 20.04: `libgtk-3-0 libwebkit2gtk-4.0-37`
- Fedora 40+: `gtk3 webkit2gtk4.1`
- RHEL / AlmaLinux / Rocky / CentOS 8-9: `gtk3 webkit2gtk3`
- Arch / Manjaro: `gtk3 webkit2gtk-4.1`

### Building from Source
If you are a developer and want to build Byto yourself, please check our [Contribution Guide](CONTRIBUTING.md).

## Usage

1.  **Launch Byto**.
2.  **Dependencies Check**: On first run, Byto will check for [yt-dlp](https://github.com/yt-dlp/yt-dlp) and [ffmpeg](https://www.ffmpeg.org/). If missing, simply click the "Download" or "Update" prompts to install them automatically.
3.  **Add Downloads**:
    - Paste a video URL into the input field.
    - (Optional) Select a specific download path.
    - Click **Add** to queue the video.
4.  **Manage Queue**:
    - Click **Start Downloads** to begin processing the queue.
    - Use settings to adjust the number of parallel downloads.
5.  **View Progress**: Watch the progress bars and status updates in real-time.

> **Note**: Make sure that all dependencies are downloaded before starting. If something downloads keep failing, check for updates in the settings.

## Contributing

Contributions are welcome! If you have ideas for new features or have found a bug, please check out our [Contribution Guide](CONTRIBUTING.md) to get started.

## License

This project is released into the public domain under the [The Unlicense](LICENSE).

---

### Disclaimer
**byto** is a neutral graphical interface built to simplify interaction with [yt-dlp](https://github.com/yt-dlp/yt-dlp). It does not promote or facilitate piracy, nor does it include any feature intended to bypass digital rights management. Users are solely responsible for ensuring that their use of this software complies with all applicable laws and with the terms of service of the platforms they access. The developers are not liable for any misuse of the software.
