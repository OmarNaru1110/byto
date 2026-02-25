# Contributing to Byto

Thank you for your interest in contributing to **Byto**! We welcome contributions from the community to help make this project better. Whether it's reporting/fixing bugs, improving documentation, or adding new features, your help is appreciated.

## ⚠️ Before You Start Developing
If you would like to work on a new feature, improvement, or refactor:
- Do not start coding immediately.
- Please open a new issue first describing what you want to work on.
- Wait for the maintainer to review it and assign it to you if it aligns with the project's goals.

This helps to:
- Avoid duplicate work.
- Ensure the feature fits the roadmap.
- Maintain consistency in design and architecture.
Only start development after the issue has been approved and assigned to you.
## Prerequisites

Before getting started, ensure you have the following installed on your machine:

- **Go**: Version 1.23 or later. [Download Go](https://go.dev/dl/)
- **Node.js**: Version 18 or later (includes npm). [Download Node.js](https://nodejs.org/)
- **Wails CLI**: Install the Wails command-line tool.
  ```bash
  go install github.com/wailsapp/wails/v2/cmd/wails@latest
  ```

## Getting Started

> **Important**: Before you start working on a feature or bug fix, please open a new issue or comment on an existing one so that I can assign it to you. This helps avoid duplicate work and ensures your contribution aligns with the project's direction.

1.  **Fork the Repository**: Click the "Fork" button on the top right of the repository page.
2.  **Clone the Repository**: Clone your forked repository to your local machine.
    ```bash
    git clone https://github.com/YOUR_USERNAME/byto.git
    cd byto
    ```

3.  **Install Dependencies**:
    - **Backend (Go)**:
      ```bash
      go mod tidy
      ```
    - **Frontend (React)**:
      ```bash
      cd frontend
      npm install
      cd ..
      ```

## Development

To start the application in development mode with live reloading:

```bash
wails dev
```

This command will:
- Build the backend
- Start the frontend dev server (Vite)
- Launch the application window

Any changes to the Go code or frontend code will automatically trigger a rebuild or reload.

## Building the Application

To build the production binary for your operating system:

```bash
wails build
```

The output binary will be located in the `build/bin` directory.

## Project Structure

- **`/app.go`**: Contains the main application logic and bridge methods exposed to the frontend.
- **`/main.go`**: Entry point of the application.
- **`/frontend`**: The React frontend application.
  - **`/src`**: Source code for the UI components.
  - **`/wailsjs`**: Auto-generated Go bindings for the frontend.
- **`/internal`**: Internal Go packages (queue, domain, builder, updater, etc.).
- **`wails.json`**: Project configuration for Wails.

## Code Style

- **Go**: Please ensure your Go code is formatted using `gofmt`.
- **Frontend**: Follow the existing coding style in the `frontend` directory.

## Submitting a Pull Request

1.  Create a new branch for your feature or bug fix:
    ```bash
    git checkout -b feature/my-new-feature
    ```
2.  Make your changes and commit them with clear, descriptive messages.
3.  Push your branch to your fork:
    ```bash
    git push origin feature/my-new-feature
    ```
4.  Open a Pull Request (PR) on the main repository.
5.  Provide a clear description of your changes and any relevant context.

## Reporting Bugs

If you encounter a bug, please open an issue on the GitHub repository with the following information:

1.  **Description**: A clear and concise description of the bug.
2.  **Steps to Reproduce**: Detailed steps to reproduce the issue.
3.  **Expected Behavior**: What you expected to happen.
4.  **Actual Behavior**: What actually happened.
5.  **Environment**: Your operating system, Go version, Node.js version, and any other relevant details.
6.  **Screenshots/Logs**: If applicable, add screenshots or error logs to help explain the problem.

## Feature Requests

Have an idea for a new feature? Open an issue and describe your suggestion. Please include why you think this feature would be useful and how it should work.
