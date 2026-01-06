# Blockchain Data & Uninstall FAQ

## The Most Important Thing to Know

**Your blockchain data lives on the blockchain, not on your computer.**

Uninstalling CLI tools is like deleting your web browser - it doesn't delete your email, bank account, or social media profiles. Those live on servers, not in your browser.

Similarly:

- Your SUI balance lives on the Sui blockchain
- Your deployed sites live on Walrus
- Your NFTs and objects live on-chain

**CLI tools are just interfaces to access the blockchain.**

## Common Questions

### Q: If I uninstall Sui CLI, will I lose my SUI balance?

**A: NO.** Your SUI balance is stored on the Sui blockchain, not on your computer.

Think of it like this:

- Sui CLI = Your web browser
- Your SUI = Your bank account
- Uninstalling Sui CLI = Closing your browser

Your bank account doesn't disappear when you close your browser!

### Q: If I uninstall Walgo, will my deployed sites disappear?

**A: NO.** Your sites are deployed on Walrus (a decentralized storage network).

Your sites will continue to be accessible at their URLs even if you:

- Uninstall Walgo
- Delete all files from your computer
- Switch to a different computer

### Q: If I uninstall site-builder, will I lose my NFTs?

**A: NO.** NFTs are stored on the Sui blockchain, not on your computer.

### Q: So what COULD I lose?

**A: Your wallet keys (if you delete ~/.sui/ without backup).**

Here's what's stored locally:

- `~/.sui/sui_config/` - Your wallet private keys
- `~/.walgo/` - Walgo cache and local database (safe to delete)
- `~/.config/walgo/` - Walgo configuration (safe to delete)

**The ONLY critical thing is `~/.sui/sui_config/`** because it contains your wallet keys.

### Q: Does Walgo uninstall delete my wallet?

**A: NO.** Walgo uninstall specifically avoids deleting `~/.sui/`.

When you run `walgo uninstall`, it removes:

- ✓ Walgo CLI binary
- ✓ Walgo Desktop app
- ✓ Walgo cache (`~/.walgo/`)

But it **NEVER** touches:

- ✗ Your Sui wallet directory (`~/.sui/`)
- ✗ Your blockchain data (it can't - it's on-chain!)

### Q: What happens if I manually delete ~/.sui/?

If you delete `~/.sui/` **without backing up your keys**:

❌ You lose:

- Access to your wallet
- Ability to sign transactions
- Control of your funds

✅ You still have:

- Your SUI balance (on blockchain)
- Your deployed sites (on Walrus)
- Your NFTs (on blockchain)

But you can't access them without the keys!

**It's like losing your house key - the house still exists, but you can't get in.**

### Q: How do I backup my wallet before uninstalling?

```bash
# 1. Get your wallet address
sui client active-address

# 2. Export your private key (save this somewhere safe!)
sui keytool export --key-identity <your-address>

# 3. Or better - save your seed phrase (if you have it)
# This was shown when you first created the wallet
```

### Q: Can I access my funds from a different computer?

**A: YES!** You have three options:

**Option 1: Copy ~/.sui/ directory**

```bash
# On old computer
tar -czf sui-wallet-backup.tar.gz ~/.sui

# Transfer to new computer, then:
tar -xzf sui-wallet-backup.tar.gz -C ~/
```

**Option 2: Use your seed phrase**

```bash
# On new computer
sui keytool import <your-seed-phrase> ed25519
```

**Option 3: Use your private key**

```bash
# On new computer
sui keytool import <your-private-key> ed25519
```

### Q: I deployed a site. Where is it stored?

Your site is stored in **three places**:

1. **Walrus Network** (decentralized storage)

   - The actual site files
   - Accessible to anyone with the URL

2. **Sui Blockchain** (metadata)

   - Site object information
   - Your ownership record
   - Configuration data

3. **Your Computer** (optional local copy)
   - `public/` folder - your built site
   - `.walgo/` cache - build cache
   - These are just for convenience, can be deleted

Deleting files on your computer doesn't affect 1 or 2!

### Q: What if I lose my Sui wallet keys?

If you lose your keys AND don't have a backup:

- ❌ You permanently lose access to that wallet
- ❌ You can't access funds in that wallet
- ❌ You can't update sites deployed from that wallet

This is fundamental to blockchain:

- No "forgot password" option
- No customer support can help
- Keys = absolute ownership

**Always backup your seed phrase!**

### Q: How do I check what I have on-chain?

```bash
# Check SUI and WAL token balances
sui client balance

# List all objects you own
sui client objects

# Get specific object details
sui client object <object-id>

# Check active address
sui client active-address
```

### Q: After reinstalling, will my sites still work?

**A: YES!** Your sites are on Walrus, not on your computer.

After reinstalling:

1. Your sites continue to be accessible at their URLs
2. If you have the same wallet, you can update them
3. If different wallet, you can still view but not update

## What Each Tool Does

### Sui CLI

**What it is:** Interface to interact with Sui blockchain
**Stores locally:** Wallet keys in `~/.sui/`
**Stores on-chain:** Your balance, objects, transactions

**Uninstalling removes:**

- CLI binary
- Ability to sign transactions (temporarily)

**Does NOT remove:**

- Your blockchain data
- Your wallet (if `~/.sui/` exists)

### Walgo

**What it is:** Tool to deploy Hugo sites to Walrus
**Stores locally:** Cache, projects database, AI history
**Stores on-chain:** Deployed site metadata

**Uninstalling removes:**

- Walgo CLI binary
- Walgo Desktop app
- Local cache and database

**Does NOT remove:**

- Deployed sites on Walrus
- Site metadata on Sui blockchain
- Your Sui wallet

### site-builder

**What it is:** Low-level tool to publish to Walrus
**Stores locally:** Nothing (stateless)
**Stores on-chain:** Site objects and blobs

**Uninstalling removes:**

- site-builder binary

**Does NOT remove:**

- Published sites
- Site objects on blockchain

## Quick Reference

| Action            | Loses Blockchain Data? | Loses Wallet Keys? | Loses Local Cache?           |
| ----------------- | ---------------------- | ------------------ | ---------------------------- |
| Uninstall Walgo   | ❌ NO                  | ❌ NO              | ✅ YES (if not --keep-cache) |
| Uninstall Sui CLI | ❌ NO                  | ❌ NO              | ❌ NO                        |
| Delete ~/.walgo/  | ❌ NO                  | ❌ NO              | ✅ YES                       |
| Delete ~/.sui/    | ❌ NO                  | ✅ YES!            | ❌ NO                        |

## Best Practices

### Before Uninstalling Any Blockchain Tool

1. ✅ Backup your seed phrase (write it down!)
2. ✅ Export your private keys
3. ✅ Verify your wallet address
4. ✅ Note down your deployed site URLs
5. ✅ Optional: Take screenshots of important data

### Safe to Delete Anytime

- `~/.walgo/` - Just cache, regenerates automatically
- `~/.config/walgo/` - Just config, can recreate
- Local site files (`public/`, etc.) - Rebuild with Hugo

### NEVER Delete Without Backup

- `~/.sui/sui_config/` - Contains wallet keys!
- Your seed phrase backup file

## Recovery Scenarios

### Scenario 1: Accidentally Deleted Walgo

```bash
# Reinstall
curl -fsSL https://raw.githubusercontent.com/selimozten/walgo/main/install.sh | bash

# Everything still works!
walgo status
```

### Scenario 2: Accidentally Deleted ~/.walgo/

```bash
# No problem! Just rebuild
cd my-site
walgo build

# Cache regenerates automatically
```

### Scenario 3: Accidentally Deleted ~/.sui/

```bash
# If you have seed phrase - you're saved!
sui keytool import <seed-phrase> ed25519

# If you have private key export - you're saved!
sui keytool import <private-key> ed25519

# If you have neither - funds are PERMANENTLY lost :(
```

### Scenario 4: Deleted Everything, New Computer

```bash
# 1. Reinstall tools
curl -fsSL https://raw.githubusercontent.com/selimozten/walgo/main/install.sh | bash

# 2. Restore wallet with seed phrase
sui keytool import <seed-phrase> ed25519

# 3. Your sites are still live!
# Just visit the URL - they never went away

# 4. Want to update a site? Rebuild locally
git clone <your-site-repo>
cd my-site
walgo build
walgo deploy
```

## The Golden Rule

**If it's important and on your computer, back it up.**
**If it's on the blockchain, it's already backed up (by the network).**

Your blockchain data is replicated across thousands of nodes worldwide. It's safer than your hard drive!

## Still Worried?

Test it yourself:

```bash
# 1. Deploy a test site
walgo init test-site
cd test-site
walgo build
walgo deploy-http --publisher https://publisher.walrus-testnet.walrus.space --aggregator https://aggregator.walrus-testnet.walrus.space --epochs 1

# 2. Note the URL
# 3. Uninstall walgo
walgo uninstall --all --force

# 4. Visit the URL - still works!
# 5. Reinstall walgo
curl -fsSL https://raw.githubusercontent.com/selimozten/walgo/main/install.sh | bash

# 6. Everything is back, site still live!
```

## Summary

🔑 **Key Point:** Blockchain data lives on the blockchain, not your computer.

✅ Safe to uninstall CLI tools anytime
✅ Your funds stay on-chain
✅ Your sites stay on Walrus
❌ Just don't delete ~/.sui/ without backup!

When in doubt: **Backup your wallet keys!**
