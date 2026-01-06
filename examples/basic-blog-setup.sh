#!/bin/bash

# Walgo Basic Blog Setup Example
# This script demonstrates a complete workflow for creating and deploying a blog with Walgo

set -e  # Exit on any error

echo "🚀 Walgo Basic Blog Setup Example"
echo "================================="

# Configuration
SITE_NAME="my-awesome-blog"
DOMAIN_NAME="awesomeblog"  # Your desired SuiNS domain (without .sui)

echo "📋 Configuration:"
echo "   Site Name: $SITE_NAME"
echo "   Domain: $DOMAIN_NAME.sui"
echo ""

# Step 1: Initialize the site
echo "1️⃣ Initializing Hugo site with Walrus configuration..."
walgo init "$SITE_NAME"
cd "$SITE_NAME"

# Step 2: Create some sample content
echo "2️⃣ Creating sample content..."

# Welcome post
walgo new posts/welcome-to-my-blog.md
cat > content/posts/welcome-to-my-blog.md << 'EOF'
---
title: "Welcome to My Blog"
date: 2024-12-19T10:00:00Z
draft: false
tags: ["welcome", "first-post"]
categories: ["general"]
---

# Welcome to My Decentralized Blog!

This is my first post on my new blog deployed to Walrus Sites - the decentralized web storage network.

## Why Decentralized?

- **Censorship resistant**: No single point of failure
- **Always available**: Distributed across multiple nodes
- **Permanent storage**: Content preserved for specified epochs
- **Cost effective**: Pay only for the storage epochs you need

## What's Next?

I'll be posting about:
- Web3 technologies
- Hugo static site generation
- Decentralized storage solutions
- And much more!

Stay tuned for more content!
EOF

# About page
walgo new about.md
cat > content/about.md << 'EOF'
---
title: "About"
date: 2024-12-19T10:00:00Z
draft: false
---

# About This Blog

This blog is built with [Hugo](https://gohugo.io) and deployed to [Walrus Sites](https://docs.walrus.site) using [Walgo](https://github.com/selimozten/walgo).

## The Technology Stack

- **Hugo**: Fast static site generator
- **Walrus Sites**: Decentralized storage network
- **SuiNS**: Decentralized naming service
- **Walgo**: CLI tool that makes it all work together seamlessly

## Contact

You can find me on the decentralized web!
EOF

# Step 3: Customize configuration
echo "3️⃣ Customizing Hugo configuration..."
cat > hugo.toml << EOF
baseURL = 'https://$DOMAIN_NAME.wal.app'
languageCode = 'en-us'
title = 'My Awesome Blog'

[params]
  description = "A blog about web3, decentralization, and technology"
  author = "Your Name"
  keywords = ["blog", "web3", "decentralization", "walrus", "hugo"]

[markup]
  [markup.goldmark]
    [markup.goldmark.renderer]
      unsafe = true

[menu]
  [[menu.main]]
    name = "Home"
    url = "/"
    weight = 10
  [[menu.main]]
    name = "Posts"
    url = "/posts/"
    weight = 20
  [[menu.main]]
    name = "About"
    url = "/about/"
    weight = 30
EOF

# Step 4: Build the site
echo "4️⃣ Building the site..."
walgo build

echo "5️⃣ Testing locally (optional)..."
echo "   Run 'walgo serve' to test your site at http://localhost:1313"
echo "   Press Ctrl+C to stop the server when ready to deploy"
echo ""

# Step 5: Deploy to Walrus
echo "6️⃣ Deploying to Walrus Sites..."
echo "   Using the interactive launch wizard..."

# Use the launch wizard for deployment
if ! walgo launch; then
    echo ""
    echo "❌ Deployment failed. This might be because:"
    echo "   1. site-builder is not installed"
    echo "   2. site-builder is not configured"
    echo "   3. You don't have sufficient funds for deployment"
    echo ""
    echo "📝 To fix this:"
    echo "   1. Install dependencies: walgo setup-deps"
    echo "   2. Set up configuration: walgo setup --network testnet"
    echo "   3. Ensure you have SUI tokens for gas fees"
    echo "   4. Try deployment again: walgo launch"
    exit 1
fi

echo ""
echo "🎉 Deployment successful!"
echo ""
echo "📝 Next steps:"
echo "   1. Save the Object ID from the deployment output to walgo.yaml"
echo "   2. Configure your SuiNS domain:"
echo "      walgo domain $DOMAIN_NAME"
echo "   3. Check your site status:"
echo "      walgo status"
echo "   4. Update your site when you add new content:"
echo "      walgo build && walgo update"
echo ""
echo "🌐 Your site will be available at:"
echo "   - https://$DOMAIN_NAME.wal.app (after SuiNS domain configuration)"
echo ""
echo "✅ Blog setup complete! Happy blogging on the decentralized web! 🚀" 