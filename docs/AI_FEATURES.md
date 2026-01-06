# AI-Powered Content Generation

## Overview

Walgo includes powerful AI features to help you generate and update Hugo content automatically. Use AI to:

- **Generate new blog posts** with proper Hugo formatting
- **Create pages** (About, Contact, Services, etc.)
- **Update existing content** with AI-powered edits
- **Maintain proper Hugo frontmatter** and Markdown structure

**Supported AI Providers:**

- **OpenAI**: GPT-4, GPT-3.5 Turbo, and other OpenAI models
- **OpenRouter**: Access to multiple models including Claude, GPT-4, Gemini, and more

## Quick Start

### 1. Configure AI Provider

First, set up your AI provider credentials:

```bash
walgo ai configure
```

You'll be prompted to:

1. Select provider (OpenAI or OpenRouter)
2. Enter your API key
3. Optionally specify a custom base URL

**Getting API Keys:**

- **OpenAI**: Get your API key at [platform.openai.com/api-keys](https://platform.openai.com/api-keys)
- **OpenRouter**: Get your API key at [openrouter.ai/keys](https://openrouter.ai/keys)

**Credentials Storage:**

- Credentials are stored securely in `~/.walgo/ai-credentials.yaml`
- File permissions are set to 0600 (owner read/write only)
- AI configuration is added to your project's `walgo.yaml`

### 2. Generate Content

Generate new blog posts or pages:

```bash
# Interactive mode
walgo ai generate

# Generate a blog post
walgo ai generate --type post

# Generate a page
walgo ai generate --type page
```

### 3. AI Pipeline (Create Complete Site)

Generate an entire site with AI:

```bash
walgo ai pipeline
```

This command:

1. Creates a site structure plan (JSON)
2. Generates each page sequentially
3. Saves progress to `.walgo/plan.json`
4. Can be resumed if interrupted

**Workflow:**

```
🤖 Enter site details:
   Site type: Blog
   Description: A personal tech blog
   Audience: Developers and tech enthusiasts

📋 Planning structure...
   ✅ Site plan saved to .walgo/plan.json

📝 Generating content...
   📄 Creating: content/posts/first-post.md
   📄 Creating: content/posts/about.md
   📄 Creating: content/contact.md

✨ Applying fixes and validation...
✅ Site generated successfully!
```

**Supported Site Types:**

- Blog - Classic blog with posts and about page
- Portfolio - Project showcase
- Docs - Documentation site
- Business - Business website

### 4. AI Plan (Create Site Plan Only)

Create a site plan without generating content:

```bash
walgo ai plan
```

This command:

- Creates a site structure plan using AI
- Saves plan to `.walgo/plan.json`
- Does not generate any content
- Plan can be reviewed before generation

**Use `walgo ai resume`** to generate content from the plan.

### 5. AI Resume (Resume from Plan)

Resume content generation from an existing plan:

```bash
walgo ai resume
```

This command:

- Loads existing `.walgo/plan.json`
- Continues generating remaining pages
- Applies post-pipeline fixes and validation
- Creates draft project entry

**Use Case:** When `walgo ai pipeline` was interrupted, you can resume from where it left off.

### 6. AI Fix (Validate and Fix Content)

Fix content files for theme requirements:

```bash
walgo ai fix
walgo ai fix --validate
```

This command:

- Validates and fixes Hugo content files
- Ensures theme-specific requirements are met
- Checks frontmatter syntax and completeness
- Removes duplicate H1 headings
- Sets proper draft status

**Flags:**

- `--validate` - Only validate, don't fix (default: false)

**For Business/Portfolio Sites (Ananke theme):**

- Validates title and description fields
- Removes duplicate H1 headings (title is already in frontmatter)
- Ensures services/projects have date fields (ISO 8601 format)
- Sets draft to false

### 6. Update Existing Content

**Example Session:**

```
🤖 AI Content Generator

Content type:
  1) Blog post
  2) Page

Select [1]: 1

Describe what content you want to generate:
> Write a blog post about deploying static sites to Walrus with Hugo

Additional context (optional, press Enter to skip):
> Focus on the benefits of decentralized hosting and how Walrus Sites work

🔄 Generating content...

✅ Content generated successfully!

Enter filename (without extension): deploying-to-walrus

✅ Content saved to: content/posts/deploying-to-walrus.md
```

### 3. Update Existing Content

Use AI to update existing Hugo files:

```bash
walgo ai update content/posts/my-post.md
```

**Example Session:**

```
🤖 AI Content Updater
📄 File: content/posts/my-post.md

Describe what changes you want to make:
> Add a new section about performance optimization and update the conclusion

🔄 Updating content...

✅ Content updated!

Save changes to file? [Y/n]: Y

✅ File updated: content/posts/my-post.md
```

## Configuration

### AI Configuration in walgo.yaml

After running `walgo ai configure`, your `walgo.yaml` will include:

```yaml
ai:
  enabled: true
  provider: openai # or "openrouter"
  model: gpt-4 # or "gpt-3.5-turbo", "anthropic/claude-3-opus", etc.
```

### Available Models

**OpenAI Models:**

- `gpt-4` - Most capable, best for complex content
- `gpt-4-turbo` - Faster, cost-effective alternative
- `gpt-3.5-turbo` - Fast and economical

**OpenRouter Models:**

- `anthropic/claude-3-opus` - Anthropic's most capable model
- `anthropic/claude-3-sonnet` - Balanced performance
- `openai/gpt-4` - GPT-4 via OpenRouter
- `google/gemini-pro` - Google's Gemini Pro
- Many more available at [openrouter.ai/models](https://openrouter.ai/models)

### Credentials Management

**View credentials location:**

```bash
ls ~/.walgo/ai-credentials.yaml
```

**Credentials file structure:**

```yaml
providers:
  openai:
    provider: openai
    api_key: sk-...
    base_url: https://api.openai.com/v1
  openrouter:
    provider: openrouter
    api_key: sk-or-...
    base_url: https://openrouter.ai/api/v1
```

**Security:**

- Never commit `ai-credentials.yaml` to version control
- File is stored in your home directory (`~/.walgo/`)
- File permissions are restrictive (0600)
- API keys are never stored in project files

## Features

### Content Generation

AI-generated content includes:

**Proper Hugo Frontmatter:**

```yaml
---
title: "Your Content Title"
date: 2025-01-15T10:30:00+00:00
draft: true
description: "SEO-friendly description"
tags: ["hugo", "static-sites", "walrus"]
categories: ["tutorials"]
---
```

**Well-Structured Markdown:**

- Clear heading hierarchy
- Proper formatting
- SEO optimization
- Engaging introductions and conclusions
- Code blocks and examples when relevant

**Content Types:**

1. **Blog Posts** (`--type post`)

   - Date-based content
   - Tags and categories
   - SEO-optimized
   - Engaging narratives

2. **Pages** (`--type page`)
   - Standalone content
   - No dates/categories
   - Professional structure
   - Informative and concise

### Content Updates

When updating existing content, AI will:

- Preserve original frontmatter (unless changes requested)
- Maintain existing style and tone
- Keep proper Hugo formatting
- Update dates if content significantly changes
- Apply requested modifications accurately

### AI System Prompts

Walgo uses specialized system prompts that understand:

- Hugo content file structure
- YAML frontmatter requirements
- Markdown best practices
- SEO optimization
- Content type differences (posts vs pages)

## Workflows

### Complete Content Creation Workflow

```bash
# 1. Configure AI (one-time setup)
walgo ai configure

# 2. Generate content
walgo ai generate --type post

# 3. Review generated content
cat content/posts/your-new-post.md

# 4. Edit if needed (manual or AI)
# Manual: edit with your preferred editor
# AI: walgo ai update content/posts/your-new-post.md

# 5. Build and preview
walgo build
# Select option 1 to preview

# 6. Deploy
walgo launch
```

### Bulk Content Generation

Generate multiple posts quickly:

```bash
# Generate first post
walgo ai generate --type post
# Instruction: "Write about Hugo basics"

# Generate second post
walgo ai generate --type post
# Instruction: "Write about Walrus Sites benefits"

# Generate third post
walgo ai generate --type post
# Instruction: "Write a tutorial on deploying with Walgo"
```

### Content Refresh Workflow

Update all existing content:

```bash
# Update each post
walgo ai update content/posts/old-post-1.md
# Instruction: "Add current best practices and update examples"

walgo ai update content/posts/old-post-2.md
# Instruction: "Improve SEO and add more details"

# Build and deploy
walgo build
walgo launch
```

## Advanced Usage

### Custom Base URLs

Use custom API endpoints:

```bash
walgo ai configure
# When prompted for base URL, enter custom endpoint
# Example: http://localhost:8000/v1 (for local models)
```

### Multiple Providers

Configure multiple providers and switch between them:

1. Configure OpenAI:

   ```bash
   walgo ai configure
   # Select OpenAI, enter key
   ```

2. Configure OpenRouter:

   ```bash
   walgo ai configure
   # Select OpenRouter, enter key
   ```

3. Switch providers in `walgo.yaml`:
   ```yaml
   ai:
     provider: openrouter # Change to "openai" to switch
     model: anthropic/claude-3-opus
   ```

### Content Templates

AI understands Hugo content patterns. You can request specific structures:

```
Instruction: "Create a tutorial-style blog post about X with:
- Introduction explaining why it matters
- Prerequisites section
- Step-by-step instructions with code examples
- Troubleshooting section
- Conclusion with next steps"
```

## Best Practices

### 1. Be Specific with Instructions

**Good:**

```
Write a comprehensive blog post about deploying static sites to Walrus.
Include sections on: benefits of decentralized hosting, step-by-step
deployment guide, cost comparison, and best practices.
```

**Less Effective:**

```
Write about Walrus
```

### 2. Provide Context

Use the "Additional context" field to provide:

- Target audience information
- Tone preferences (technical, casual, professional)
- Specific examples to include
- Length preferences

### 3. Review AI Output

Always review and edit AI-generated content:

- Check technical accuracy
- Verify links and references
- Adjust tone and style
- Add personal insights
- Update dates and metadata

### 4. Iterate with Updates

If first generation isn't perfect:

```bash
# Generate initial content
walgo ai generate

# Update with refinements
walgo ai update content/posts/file.md
# Instruction: "Make the introduction more engaging and add code examples"
```

### 5. Set Appropriate Draft Status

- AI sets `draft: true` by default
- Review content before setting to `false`
- Use Hugo's draft preview: `hugo server -D`

## Troubleshooting

### "AI features are not enabled"

**Solution:**

```bash
walgo ai configure
```

### "No credentials found for provider"

**Solution:**

```bash
# Reconfigure credentials
walgo ai configure

# Or manually check credentials file
cat ~/.walgo/ai-credentials.yaml
```

### "API request failed"

**Possible causes:**

1. Invalid API key - verify your key is correct
2. Insufficient credits/quota - check your provider account
3. Network issues - check internet connection
4. Rate limiting - wait and try again

**Solution:**

```bash
# Reconfigure with valid credentials
walgo ai configure
```

### "Error loading credentials"

**Solution:**

```bash
# Check credentials file exists
ls ~/.walgo/ai-credentials.yaml

# If missing, reconfigure
walgo ai configure
```

### Generated Content Has Issues

**Common issues:**

- Missing frontmatter fields - update instructions to be more specific
- Wrong date format - regenerate or manually fix
- Incorrect content type - specify `--type` flag

**Solution:**

```bash
# Regenerate with better instructions
walgo ai generate --type post

# Or update the problematic file
walgo ai update content/posts/file.md
# Instruction: "Fix frontmatter date format to ISO 8601"
```

## Costs

### OpenAI Pricing

Approximate costs (as of 2025):

- GPT-4: ~$0.03 per 1K tokens (input), ~$0.06 per 1K tokens (output)
- GPT-3.5 Turbo: ~$0.0015 per 1K tokens (input), ~$0.002 per 1K tokens (output)

**Typical content generation:**

- Blog post (500-1000 words): $0.10-$0.30 with GPT-4
- Short update: $0.05-$0.10 with GPT-4

### OpenRouter Pricing

Varies by model:

- Claude 3 Opus: ~$15 per 1M input tokens
- GPT-4: Similar to OpenAI pricing
- Check current pricing: [openrouter.ai/models](https://openrouter.ai/models)

**Benefits of OpenRouter:**

- Access to multiple providers with one API key
- Automatic fallback between models
- Unified billing

## Security & Privacy

### Credentials

- API keys stored locally in `~/.walgo/ai-credentials.yaml`
- File permissions: 0600 (owner read/write only)
- Never stored in project files or version control
- Never transmitted except to configured AI provider

### Content Privacy

When you generate content:

- Your instructions are sent to the AI provider
- Generated content is returned to your local machine
- Content is stored locally in your Hugo site
- No data is stored on Walgo servers (there are none!)

**Important:**

- Review your AI provider's privacy policy
- Don't include sensitive information in prompts
- Be aware that prompts may be used for model training (provider-dependent)

## Examples

### Example 1: Generate Tutorial Post

```bash
walgo ai generate --type post
```

**Instruction:**

```
Create a detailed tutorial on setting up a Hugo blog with the Ananke theme.
Include installation steps, configuration examples, and how to create your
first post. Target audience: developers new to Hugo.
```

**Result:**
A complete blog post with:

- Clear installation instructions
- Configuration examples with code blocks
- Step-by-step first post creation
- Proper frontmatter and SEO

### Example 2: Create About Page

```bash
walgo ai generate --type page
```

**Instruction:**

```
Create an About page for a software developer's portfolio site. Include
sections for: background, skills, experience, and contact information.
Keep it professional and concise.
```

**Result:**
A well-structured About page with appropriate frontmatter and professional content.

### Example 3: Update Existing Post

```bash
walgo ai update content/posts/walrus-intro.md
```

**Instruction:**

```
Add a new section called "Cost Comparison" comparing Walrus storage costs
to traditional cloud hosting. Update the conclusion to reference this new section.
```

**Result:**
Original post updated with new section, maintaining existing structure and style.

### Example 4: Refresh Old Content

```bash
walgo ai update content/posts/old-tutorial.md
```

**Instruction:**

```
This tutorial is from 2023. Update it to reflect 2025 best practices,
update any deprecated commands, and improve code examples. Keep the
same overall structure and tone.
```

**Result:**
Refreshed content with current information while preserving original intent.

## FAQ

**Q: Which AI provider should I use?**
A:

- **OpenAI**: Best for straightforward use, well-documented, reliable
- **OpenRouter**: Best for accessing multiple models (Claude, GPT-4, etc.) with one API key

**Q: Can I use local AI models?**
A: Yes! Configure a custom base URL pointing to your local model server (must be OpenAI API-compatible).

**Q: Will AI always generate perfect content?**
A: No. AI-generated content should always be reviewed and edited. Think of it as a first draft.

**Q: How much does it cost?**
A: Varies by provider and model. Typical blog post: $0.05-$0.30. See Costs section above.

**Q: Can AI update my Hugo theme files?**
A: The AI features are designed for content (markdown files), not theme files. Manual editing recommended for themes.

**Q: Is my API key secure?**
A: Yes. It's stored locally with restrictive permissions and never committed to version control.

**Q: Can I generate content in other languages?**
A: Yes! Simply write your instructions in your target language, and AI will generate content in that language.

**Q: What if I don't like the generated content?**
A: You can:

1. Regenerate with better instructions
2. Use `walgo ai update` to refine it
3. Manually edit the file
4. Delete it and start over

**Q: Can I batch process multiple files?**
A: Currently, each command processes one file. For batch updates, write a shell script that calls `walgo ai update` for each file.

## Related Commands

- `walgo init` - Create new Hugo site
- `walgo new` - Create new content (traditional way)
- `walgo build` - Build site
- `walgo serve` - Preview site locally
- `walgo launch` - Deploy to Walrus

## Support

For issues or questions:

- GitHub Issues: [github.com/selimozten/walgo/issues](https://github.com/selimozten/walgo/issues)
- Check provider documentation: [OpenAI](https://platform.openai.com/docs) | [OpenRouter](https://openrouter.ai/docs)
