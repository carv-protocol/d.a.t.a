# D.A.T.A (Data Authentication, Trust, and Attestation) Framework

[![License: ISC](https://img.shields.io/badge/License-ISC-blue.svg)](https://opensource.org/licenses/ISC)

> An AI-native framework for autonomous blockchain data interaction and analysis, developed by CARV.

## 🌟 Overview

D.A.T.A is a cutting-edge framework that bridges the gap between AI agents and blockchain data. Built as a plugin for AI frameworks like Eliza, it enables AI agents to autonomously fetch, process, and act on both on-chain and off-chain data, fostering intelligent, data-driven decision-making in the Web3 space.

## 📖 Documentation

- [Documentation](https://docs.carv.io/d.a.t.a.-ai-framework/introduction)

## 🚀 Features

- **On-Chain Data Access**
  - Real-time blockchain data fetching (transactions, balances, activity metrics)
  - Multi-chain support (Ethereum, Base, Bitcoin Solana)
  - Transaction monitoring and analysis
  - Smart contract interaction

- **Off-Chain Data Integration**
  - User profiles and behavioral analytics
  - Token metadata and market information
  - Contextual data enrichment

- **AI-Native Architecture**
  - Vector database integration for semantic search
  - Autonomous decision-making capabilities
  - Built-in memory management for AI agents

- **Cross-Chain Analytics**
  - Unified data aggregation across blockchains
  - Standardized query interface
  - Comprehensive blockchain activity analysis

- **Social Integration**
  - Twitter integration
  - Discord support
  - Telegram bot functionality
  - Community engagement tracking

## 📋 Prerequisites

- Go 1.21 or higher
- Make
- Access to blockchain node or data provider
- API credentials for social platforms

## 🔧 Installation

1. Clone the repository:
```bash
git clone https://github.com/carv-protocol/d.a.t.a.git
cd d.a.t.a
```

2. Install dependencies:
```bash
make tidy
```

3. Build the project:
```bash
make build
```

### Configuration

1. Create a `.env` file in the root directory:
```env
# LLM Configuration
LLM_API_KEY=your_api_key
LLM_BASE_URL=https://api.openai.com/v1

# Social Media Tokens
TWITTER_API_KEY=your_twitter_key
DISCORD_TOKEN=your_discord_token
TELEGRAM_BOT_TOKEN=your_telegram_token

2. Add your CARV API credentials:
DATA_API_KEY=your_api_key
DATA_AUTH_TOKEN=your_auth_token

3. Update the config file:
- src/config/config.yaml
- src/config/character_*.json
```

## 📮 Contact & Support

- [GitHub Issues](https://github.com/carv-protocol/d.a.t.a/issues)
- [Discord Community](http://discord.gg/AYyfmhMn5K)
- [Documentation](https://docs.carv.io/d.a.t.a.-ai-framework/introduction)
- Email: developer@carv.io