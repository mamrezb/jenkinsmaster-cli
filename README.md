
# JenkinsMaster CLI 🚀

[![GitHub Release](https://img.shields.io/github/v/release/mamrezb/jenkinsmaster-cli?color=brightgreen&style=for-the-badge)](https://github.com/mamrezb/jenkinsmaster-cli/releases)
[![License](https://img.shields.io/github/license/mamrezb/jenkinsmaster-cli?color=blue&style=for-the-badge)](https://github.com/mamrezb/jenkinsmaster-cli/blob/main/LICENSE)
[![Contributions Welcome](https://img.shields.io/badge/contributions-welcome-orange?style=for-the-badge)](https://github.com/mamrezb/jenkinsmaster-cli/issues)

JenkinsMaster CLI empowers developers to deploy and manage production-ready Jenkins instances with ease. A tool crafted with precision, it integrates Terraform and Ansible to deliver a seamless CI/CD experience. 🛠️

---

## 🌟 Features
- **Interactive Deployments**: Guided setup through intuitive CLI prompts.
- **Cloud & SSH Support**: Deploy on Hetzner Cloud or existing infrastructure via SSH.
- **Automation**: Provision with Terraform and configure Jenkins using Ansible.
- **Custom Jenkins**: Full control over credentials, plugins, and configurations.
- **Modular Design**: Extend and adapt as your needs evolve.

---

## 🔧 Prerequisites
Ensure you have the following:
- **Git**
- **Ansible** ([Installation Guide](https://docs.ansible.com/ansible/latest/installation_guide/intro_installation.html))
- **Terraform** ([Installation Guide](https://learn.hashicorp.com/terraform/getting-started/install.html))
- **SSH Key** for secure server access.

---

## 📥 Installation

### 🏗️ Via Homebrew (macOS/Linux)
```bash
brew tap mamrezb/jenkinsmaster-cli
brew install jenkinsmaster
```

### 📦 Binary Releases
Download binaries for your platform from the [Releases](https://github.com/mamrezb/jenkinsmaster-cli/releases) page.

### 🛠️ Build from Source
```bash
git clone https://github.com/mamrezb/jenkinsmaster-cli.git
cd jenkinsmaster-cli
go build -o jenkinsmaster
```

---

## 🚀 Getting Started

Start your deployment journey:
```bash
jenkinsmaster deploy
```

1. Choose a provider: **Hetzner Cloud** or **SSH**.
2. Follow the interactive prompts for credentials and configurations.
3. Let the magic happen! ✨ JenkinsMaster CLI handles everything from infrastructure to Jenkins setup.

---

## 🔌 Key Repositories
- **Ansible Role**: [jenkinsmaster-ansible-role](https://github.com/mamrezb/jenkinsmaster-ansible-role)
- **Terraform Module**: [terraform-hcloud-jenkinsmaster](https://github.com/mamrezb/terraform-hcloud-jenkinsmaster)
- **Job DSL**: [jenkinsmaster-job-dsl](https://github.com/mamrezb/jenkinsmaster-job-dsl)
- **Shared Libraries**: [jenkinsmaster-shared-library](https://github.com/mamrezb/jenkinsmaster-shared-library)
- **Sample Backend**: [jenkinsmaster-backend-helloworld](https://github.com/mamrezb/jenkinsmaster-backend-helloworld)
- **Sample Frontend**: [jenkinsmaster-frontend-helloworld](https://github.com/mamrezb/jenkinsmaster-frontend-helloworld)

---

## 🤝 Contributions
Contributions are the ❤️ of open source! Here's how you can help:
1. Fork this repository.
2. Create a branch (`feature/super-feature`).
3. Commit your changes (`git commit -m "Add super feature"`).
4. Push to the branch (`git push origin feature/super-feature`).
5. Open a Pull Request.

For details, see the [Contributing Guide](CONTRIBUTING.md).

---

## 📄 License
This project is licensed under the [MIT License](LICENSE).

---

## ✨ Stay Connected
- **Author**: [mamrezb](https://github.com/mamrezb)
- **Email**: [behfar.mr@gmail.com](mailto:behfar.mr@gmail.com)

Mastering Jenkins has never been easier. Deploy confidently, automate seamlessly, and accelerate your development pipeline! 💪