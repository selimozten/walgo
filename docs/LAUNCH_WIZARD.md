# Launch Wizard & Project Management

## Overview

Walgo provides powerful interactive tools for deploying and managing your Walrus sites:

- **`walgo launch`** - Interactive deployment wizard
- **`walgo projects`** - Project management and history

All deployments are automatically tracked in a SQLite database for easy management and updates.

## Launch Wizard

### What is it?

`walgo launch` is an interactive step-by-step wizard that guides you through deploying your site to Walrus. Perfect for beginners and anyone who wants a guided deployment experience.

### Usage

#### Direct Launch

```bash
cd my-site
walgo launch
```

#### Build → Preview → Launch Flow

For the best experience, use the integrated build workflow:

```bash
cd my-site
walgo build
```

After building, you'll be prompted with three options:

1. **Preview site locally** - Start a local server to preview your site before deployment
2. **Launch deployment wizard** - Go directly to the launch wizard
3. **Exit** - Exit without deploying

**Preview Experience:**

- Uses Hugo's built-in server (http://localhost:1313) for the best preview with live reload
- Falls back to simple HTTP server (http://localhost:8080) if Hugo server unavailable
- After previewing, you can continue directly to the launch wizard or exit

This creates a smooth workflow: **build → preview → launch**

### Wizard Steps

#### Step 1: Choose Network

Select your deployment target:

- **Testnet**: For testing (1 epoch = 1 day)
- **Mainnet**: For production (1 epoch = 2 weeks, requires SuiNS)

The wizard automatically switches your Sui client to the selected network.

#### Step 2: Wallet & Balance

- Shows your current active wallet address
- Displays your SUI balance
- Option to switch wallets or add new addresses

**Note**: You need sufficient SUI to pay for gas fees and storage.

#### Step 3: Project Details

- **Project Name** (required): A friendly name for your site
- **Category** (optional): Organize projects (e.g., blog, portfolio, docs)

Projects are saved for easy future updates.

#### Step 4: Storage Duration

Choose how long to store your site:

**Testnet:**

- 1 epoch = 1 day
- Maximum: 53 epochs
- Suggested: 1-30 epochs (1 day - 1 month)

**Mainnet:**

- 1 epoch = 2 weeks
- Maximum: 53 epochs
- Suggested durations:
  - 2 epochs = 1 month
  - 6 epochs = 3 months
  - 26 epochs = 1 year

#### Step 5: Verify Site

The wizard checks that your site is built and ready:

- Verifies `public/` directory exists
- Calculates site size
- Shows deployment location

**If your site isn't built:**

```bash
walgo build
walgo launch
```

#### Step 6: Review & Confirm

Review your deployment summary:

- Network
- Project name and category
- Wallet address and balance
- Storage duration
- Site size
- **Estimated gas fee**

**Confirm to deploy** or cancel if you need to make changes.

#### Step 7: Deploy

The wizard deploys your site:

- Uploads files to Walrus
- Creates site object on Sui blockchain
- Records deployment in project database
- Displays Object ID and access URLs

#### Step 8: Success!

After successful deployment:

- Site Object ID displayed
- **SuiNS configuration instructions** for public access (both testnet and mainnet)
- Project saved for future management
- Next steps suggested

### Configuring SuiNS (Post-Deployment)

After deploying your site, you'll receive instructions to make it publicly accessible via SuiNS.

**Note**: B36 subdomains are no longer supported on the wal.app portal. You must use SuiNS names for public access.

For a complete guide, see the [official SuiNS tutorial](https://docs.wal.app/docs/walrus-sites/tutorial-suins).

#### Step 1: Get a SuiNS Domain

**For Mainnet:**

1. Visit [suins.io](https://suins.io)
2. Connect your Sui wallet
3. Search for your desired domain
4. Purchase the domain

**Important**: Domain names can only contain letters (a-z) and numbers (0-9). No special characters like hyphens are allowed.

#### Step 2: Link SuiNS to Your Walrus Site

1. Go to "Names You Own" section:

   - **Mainnet**: [suins.io](https://suins.io)

2. Find your domain and click the "three dots" menu icon

3. Click **"Link To Walrus Site"**

4. Paste your Site Object ID (displayed after deployment)

5. Verify the Object ID is correct

6. Click "Apply" and approve the transaction

#### Step 3: Access Your Site

Your site will now be accessible at: `https://your-domain.wal.app`

**Example**: If you registered `myportfolio`, your site is at:

- Mainnet: `https://myportfolio.wal.app`
- Testnet: `https://myportfolio.wal.app` (using testnet SuiNS)

#### Backwards Compatibility

If you previously used "Link To Wallet Address", that still works but is deprecated. We recommend using "Link To Walrus Site" for all new sites and updates.

## Project Management

### View All Projects

```bash
walgo projects
# or
walgo projects list
```

**Example output:**

```
╔═══════════════════════════════════════════════════════════╗
║                  Your Walrus Projects                     ║
╚═══════════════════════════════════════════════════════════╝

📦 My Portfolio (portfolio)
   ID:           1
   Network:      mainnet
   Status:       active
   Object ID:    0x1234...
   SuiNS:        myportfolio
   Deployments:  5
   Last deploy:  2024-01-15 10:30
   Updated:      2 days ago

─────────────────────────────────────────────────────────────

📦 Test Blog (blog)
   ID:           2
   Network:      testnet
   Status:       active
   Object ID:    0x5678...
   Deployments:  12
   Last deploy:  2024-01-20 14:22
   Updated:      3 hours ago
```

### Filter Projects

```bash
# Filter by network
walgo projects list --network mainnet
walgo projects list --network testnet

# Filter by status
walgo projects list --status active
walgo projects list --status archived
```

### Show Project Details

```bash
walgo projects show <name>
walgo projects show <id>
```

**Example:**

```bash
walgo projects show "My Portfolio"
```

**Displays:**

- Project information (name, network, status, Object ID, SuiNS)
- Statistics (total deployments, success rate, dates)
- Recent deployment history
- Available actions

### Update a Project

Redeploy an existing project with new content:

```bash
walgo projects update <name>

# With custom epochs
walgo projects update <name> --epochs 10
```

**What happens:**

1. Changes to the project's site directory
2. Checks that site is built
3. Deploys update to Walrus
4. Records new deployment in history
5. Updates project Object ID

**Example:**

```bash
cd /path/to/my-site
# Make changes to your site
walgo build
walgo projects update "My Portfolio"
```

### Archive a Project

Mark a project as archived (hides from default list):

```bash
walgo projects archive <name>
```

Archived projects:

- Not shown in default project list
- Can be restored later
- History preserved
- Site continues to exist on Walrus

**To view archived projects:**

```bash
walgo projects list --status archived
```

### Delete a Project

Permanently delete a project and its history:

```bash
walgo projects delete <name>
```

**Warning:**

- Deletes project record from database
- Deletes all deployment history
- **Does NOT** delete the site from Walrus
- Cannot be undone

The wizard prompts for confirmation before deletion.

## Database Storage

### Location

Projects are stored in a SQLite database:

```
~/.walgo/projects.db
```

### What's Stored

**Projects table:**

- Project metadata (name, category, network)
- Current deployment info (Object ID, SuiNS, epochs)
- Wallet address used
- Site path on local system
- Timestamps (created, updated, last deployed)
- Deployment count
- Status (active/archived)

**Deployments table:**

- Complete deployment history
- Object ID for each deployment
- Network and configuration
- Success/failure status
- Timestamps
- Optional version tags and notes

### Data Privacy

- All data stored locally on your machine
- No data sent to external servers
- Wallet addresses stored for reference only
- You can delete the database anytime

## Network Configuration

### Testnet

- **Epoch Duration**: 1 day
- **Maximum Epochs**: 53 (53 days)
- **SuiNS**: Optional
- **Best For**: Testing, development, short-term demos
- **Cost**: Free (testnet tokens)

### Mainnet

- **Epoch Duration**: 2 weeks
- **Maximum Epochs**: 53 (106 weeks ≈ 2 years)
- **SuiNS**: Required for public access
- **Best For**: Production sites, long-term hosting
- **Cost**: Real SUI tokens for gas and storage

## Gas Fee Estimation

The wizard estimates gas fees based on:

- Site size (larger sites cost more)
- Network (mainnet typically higher)
- Number of epochs

**Rough estimates:**

- Small site (< 1 MB): ~0.01-0.02 SUI
- Medium site (1-10 MB): ~0.02-0.05 SUI
- Large site (10+ MB): ~0.05-0.1+ SUI

**Actual fees vary** based on network conditions and exact site structure.

## Best Practices

### Project Organization

1. **Use descriptive names**: "Company Blog", "Personal Portfolio"
2. **Add categories**: Group similar projects
3. **Keep projects updated**: Regular deployments ensure fresh content
4. **Archive old projects**: Keep active list manageable

### Deployment Strategy

1. **Test on testnet first**: Always test before mainnet
2. **Use version tags**: Track significant updates
3. **Monitor gas costs**: Larger sites cost more
4. **Choose appropriate epochs**: Balance cost vs. duration

### Wallet Management

1. **Use dedicated wallet**: Consider separate wallet for deployments
2. **Keep funded**: Ensure sufficient SUI balance
3. **Track spending**: Monitor gas fees over time
4. **Backup wallet**: Never lose access to your sites

## Examples

### First Deployment

```bash
# Create and build site
walgo init my-blog
cd my-blog
walgo build

# After build, you'll see:
# What would you like to do next?
#   1) Preview site locally
#   2) Launch deployment wizard
#   3) Exit

# Select 1 to preview
# Preview opens at http://localhost:1313
# Press Enter when done

# Continue to launch wizard? [Y/n]: Y

# Follow launch wizard prompts:
# 1. Select testnet
# 2. Confirm wallet
# 3. Name: "My Blog"
# 4. Category: "blog"
# 5. Epochs: 30 (1 month)
# 6. Confirm and deploy
```

### First Deployment (Direct Launch)

```bash
# Skip preview and launch directly
walgo init my-blog
cd my-blog
walgo build

# Select option 2 (Launch deployment wizard)
# Or run launch directly:
walgo launch

# Follow prompts as above
```

### Update Existing Site

```bash
# Make changes
cd my-blog
# ... edit content ...
walgo build

# Update via projects
walgo projects update "My Blog"
```

### View Project History

```bash
# Show all details
walgo projects show "My Blog"

# View deployments, stats, and history
```

### Mainnet Production Site

```bash
# Build site
cd company-website
walgo build

# Launch wizard
walgo launch

# Follow prompts:
# 1. Select mainnet
# 2. Confirm wallet (ensure funded)
# 3. Name: "Company Website"
# 4. Category: "corporate"
# 5. Configure SuiNS: "mycompany"
# 6. Epochs: 26 (1 year)
# 7. Review gas estimate
# 8. Confirm and deploy
```

## Troubleshooting

### "No walgo.yaml found"

**Problem**: Not in a Walgo project directory

**Solution**:

```bash
walgo init my-site
cd my-site
walgo launch
```

### "Site not built"

**Problem**: No `public/` directory

**Solution**:

```bash
walgo build
walgo launch
```

### "Insufficient balance"

**Problem**: Not enough SUI for gas fees

**Solution**:

- **Testnet**: Get tokens from [faucet.sui.io](https://faucet.sui.io)
- **Mainnet**: Purchase SUI on exchange

### "SuiNS required on mainnet"

**Problem**: Mainnet deployment without SuiNS

**Options**:

1. Register SuiNS domain first
2. Deploy and configure SuiNS later (manual)
3. Use testnet for testing

### "Project not found"

**Problem**: Project doesn't exist in database

**Solution**:

```bash
# List all projects
walgo projects list

# Check exact name
walgo projects show <exact-name>
```

## Related Commands

- `walgo init` - Create new site
- `walgo build` - Build site (with integrated preview and launch options)
- `walgo launch` - Launch deployment wizard
- `walgo deploy` - Direct deployment (without wizard)
- `walgo projects` - Manage deployed projects
- `walgo status` - Check deployment status
- `walgo doctor` - Diagnose issues

## FAQ

**Q: Can I have multiple projects?**
A: Yes! Manage unlimited projects with the projects command.

**Q: Does updating change the Object ID?**
A: Yes, each deployment creates a new Object ID.

**Q: Can I deploy the same site to both networks?**
A: Yes, testnet and mainnet are separate. Create two projects.

**Q: What happens to old deployments?**
A: They remain on Walrus until epochs expire. History is preserved in database.

**Q: Can I delete my project database?**
A: Yes, but you'll lose all project history and tracking. Sites on Walrus are unaffected.

**Q: How do I backup my projects?**
A: Copy `~/.walgo/projects.db` to backup location.
