package ai

import (
	"fmt"
	"strings"
)

// =============================================================================
// PROMPT ARCHITECTURE
// =============================================================================
//
// Modular, composable prompt components:
//
// 1. OUTPUT RULES       - Format, YAML syntax
// 2. CONTENT RULES      - Quality, SEO
// 3. HUGO RULES         - Structure, linking, frontmatter
// 4. SITE PLANNER       - Planning prompts and schema
// 5. PROMPT COMPOSERS   - Combine components into full prompts
// 6. PROMPT BUILDERS    - Build user prompts with context
// 7. UTILITIES          - Content cleaning, helpers
//
// =============================================================================

// =============================================================================
// 1. OUTPUT RULES
// =============================================================================

// OutputFormatRules defines the expected output format for all content generation
const OutputFormatRules = `OUTPUT FORMAT:
- First line MUST be --- (YAML frontmatter opening)
- Frontmatter ends with --- on its own line
- Content follows immediately after frontmatter
- NEVER wrap output in code fences (no triple backticks)
- Return ONLY the complete Hugo markdown file
- NO explanations, comments, or meta-text outside the file`

// YAMLSyntaxRules defines strict YAML syntax requirements
const YAMLSyntaxRules = `YAML SYNTAX (CRITICAL):
- ALWAYS use double quotes (") for ALL string values
- NEVER use single quotes (') - they break with apostrophes
- Special characters require quoting: : ' # " [ ] { } @ & * ? | > < = !

CORRECT:
  title: "A Developer's Guide to Hugo"
  description: "Let's build something amazing: a complete guide"

WRONG (causes parse errors):
  title: 'Developer's Guide'     # apostrophe breaks single quotes
  description: A guide: complete  # unquoted colon causes error`

// =============================================================================
// 2. CONTENT RULES
// =============================================================================

// ContentQualityRules defines professional content standards
const ContentQualityRules = `CONTENT QUALITY:

VOICE:
- Write like a knowledgeable human — not a content mill, not a corporate brochure
- Be direct and specific — replace vague claims with concrete details
- Vary sentence length and structure naturally

BANNED PATTERNS (these make content feel AI-generated):
- "In today's [digital/modern/fast-paced] [world/landscape/era]"
- "Welcome to [our/my] [site/blog/company]"
- "We are passionate about..."
- "Whether you're a [X] or [Y]..."
- "In the ever-evolving world of..."
- "It's not just about X, it's about Y"
- Starting multiple paragraphs the same way
- Listicle headlines ("7 Proven Ways to...")

STRUCTURE:
- Hook readers in the first sentence — question, bold claim, or specific insight
- One idea per paragraph, 2-4 sentences max
- Use H2/H3 headers to make content skimmable
- End with a clear, specific call-to-action

SUBSTANCE:
- Replace adjectives with evidence ("fast" → "responds in 50ms")
- Show expertise through depth, not through claiming to be an expert
- Include specific examples, techniques, or data points
- Every sentence must earn its place — if removing it loses nothing, remove it`

// SEORules defines search engine optimization requirements
const SEORules = `SEO:
- Title: 50-60 chars, primary keyword first when natural
- Description: 150-160 chars with primary keyword
- H1 auto-generated from title — NEVER use # in body
- Start body with H2 (##), use H3 (###) for subsections
- Never skip heading levels
- Use descriptive anchor text for internal links
- Integrate keywords naturally — no stuffing`

// =============================================================================
// 3. HUGO RULES
// =============================================================================

// HugoRules defines Hugo content structure and linking requirements
const HugoRules = `HUGO STRUCTURE:

CONTENT ORGANIZATION:
- content/ is the root for all content
- Top-level directories = sections (posts/, docs/, projects/)
- _index.md: Section landing pages and home page (branch bundle)
- index.md: Pages with bundled images/resources (leaf bundle)
- name.md: Simple pages without resources

HEADING RULES:
- NEVER use H1 (#) in body — title generates it
- Start body with H2 (##) for main sections
- Use H3 (###) for subsections
- Never skip heading levels

INTERNAL LINKS:
- ALWAYS use absolute paths: /about/, /posts/my-post/
- ALWAYS include trailing slash
- NEVER use relative paths (../about) or ref/relref shortcodes

FRONTMATTER:
- title: required (quoted string)
- draft: false (ALWAYS)
- description: recommended for SEO
- date: ISO 8601 format for dated content
- weight: integer for ordering (docs, menus)`

// =============================================================================
// 4. SITE PLANNER
// =============================================================================

// ComposeSitePlannerPrompt creates the system prompt for site planning
func ComposeSitePlannerPrompt() string {
	return `You are a SITE ARCHITECT. Plan a Hugo site structure.
Output ONLY valid JSON. No Markdown, explanations, or comments.

` + sitePlannerPrinciples + `

` + sitePlannerPageCount + `

` + sitePlannerJSONSchema + `

` + sitePlannerSiteTypeRules + `

` + sitePlannerCriticalRules
}

const sitePlannerPrinciples = `PLANNING PRINCIPLES:
- Every page MUST have a clear purpose and reason to exist — zero filler
- Build trust: case studies, results, social proof
- User journey: Awareness → Consideration → Decision → Action
- Quality over quantity: 5 excellent pages beat 15 mediocre ones
- But don't artificially limit — if the site needs 20 pages, plan 20 pages

CONTENT QUALITY (CRITICAL):
- Every page must contain substantive, specific, useful content
- NO generic placeholder text ("Lorem ipsum", "Welcome to our site", "We are passionate about...")
- NO vague corporate speak — be specific, concrete, and valuable
- Write like a human expert, not a content mill
- Each page should stand alone as genuinely useful to the reader`

const sitePlannerPageCount = `PAGE COUNT - YOU DECIDE:
- Analyze the description, audience, and site type
- Create exactly the pages that serve a real purpose — no more, no less
- A simple site may need 3 pages, a complex one may need 20+
- NEVER pad with filler pages — every page must have a clear reason to exist
- The only hard requirement is a homepage (content/_index.md)
- About, contact, and other pages are optional — include them only if they make sense for this specific site`

const sitePlannerJSONSchema = `OUTPUT JSON SCHEMA:
{
  "site": {
    "type": "blog|docs",
    "title": "...",
    "language": "en",
    "tone": "...",
    "base_url": "..."
  },
  "pages": [
    {
      "id": "home",
      "path": "content/_index.md",
      "page_type": "home",
      "frontmatter": {
        "title": "...",
        "draft": false,
        "description": "..."
      },
      "outline": ["bullet1", "bullet2", "..."],
      "internal_links": ["/about/", "/contact/"]
    }
  ]
}`

const sitePlannerSiteTypeRules = `STRUCTURE BY SITE TYPE:

BLOG:
- content/_index.md (homepage)
- content/posts/_index.md (section index, if multiple posts)
- content/posts/*/index.md (blog posts — as many as the description warrants)
- Additional pages as needed (about, contact, etc.)

DOCS:
- content/_index.md (landing)
- content/docs/_index.md (docs root)
- content/docs/*/*.md (structured documentation — cover the full scope described)`

const sitePlannerCriticalRules = `CRITICAL RULES:
- _index.md: Required for sections (docs/, projects/, services/)
- NOT for root pages (about, contact)
- Internal links: Absolute with trailing slash (/about/, /posts/welcome/)
- NO H1 in body content
- NO tag/category/taxonomy/search/RSS pages
- draft: false always
- Return ONLY JSON`

// =============================================================================
// 5. PROMPT COMPOSERS
// =============================================================================

// ComposePageGeneratorPrompt creates the system prompt for single page generation.
// themeContext should be generated using BuildDynamicThemeContext() from theme_analyzer.go.
func ComposePageGeneratorPrompt(themeContext string) string {
	var sb strings.Builder

	sb.WriteString(`You are a WORLD-CLASS CONTENT CREATOR.
Generate ONE Hugo page using the provided page plan.

`)
	sb.WriteString(OutputFormatRules)
	sb.WriteString("\n\n")
	sb.WriteString(YAMLSyntaxRules)
	sb.WriteString("\n\n")
	sb.WriteString(HugoRules)

	if themeContext != "" {
		sb.WriteString("\n\n")
		sb.WriteString(themeContext)
	}

	sb.WriteString("\n\n")
	sb.WriteString(ContentQualityRules)
	sb.WriteString("\n\n")
	sb.WriteString(SEORules)
	sb.WriteString("\n\nReturn ONLY the complete Hugo markdown file.")

	return sb.String()
}

// ComposeContentUpdatePrompt creates the system prompt for content updates
func ComposeContentUpdatePrompt() string {
	return `You are a Hugo content editor specializing in updates.

` + OutputFormatRules + `

` + YAMLSyntaxRules + `

UPDATE RULES:
- Preserve ALL existing frontmatter fields unless asked to change
- Preserve tone, structure, and working links
- Fix YAML/Markdown syntax errors if found
- Do NOT add tags/categories unless requested
- Do NOT introduce Hugo ref/relref shortcodes
- Keep draft: false unless user explicitly requests drafts

OUTPUT:
- Return the COMPLETE updated Hugo markdown file
- Include ALL original content plus requested changes
- Start with --- and ensure valid YAML/Markdown`
}

// =============================================================================
// 6. PROMPT BUILDERS
// =============================================================================

// BuildSitePlannerPrompt builds the user prompt for site planning
func BuildSitePlannerPrompt(siteName, siteType, description, audience, tone, baseURL string) string {
	return BuildSitePlannerPromptWithTheme(siteName, siteType, description, audience, tone, baseURL, "", "")
}

// BuildSitePlannerPromptWithTheme builds the user prompt with dynamic theme analysis
func BuildSitePlannerPromptWithTheme(siteName, siteType, description, audience, tone, baseURL, sitePath, themeName string) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf(`SITE NAME: %s
SITE TYPE: %s
DESCRIPTION: %s
TARGET AUDIENCE: %s
TONE: %s
BASE URL: %s`,
		siteName, siteType, description, audience, tone, baseURL))

	// Add dynamic theme analysis if available
	if sitePath != "" && themeName != "" {
		themeAnalysis := AnalyzeTheme(sitePath, themeName)
		configAnalysis := AnalyzeThemeConfig(sitePath, themeName)

		sb.WriteString(fmt.Sprintf("\n\nTHEME: %s", themeName))

		// Supported sections from theme
		if len(themeAnalysis.Sections) > 0 {
			sb.WriteString(fmt.Sprintf("\nSUPPORTED SECTIONS: %s", strings.Join(themeAnalysis.Sections, ", ")))
		}

		// Frontmatter fields per section
		if len(themeAnalysis.FrontmatterFields) > 0 {
			sb.WriteString("\nFRONTMATTER BY SECTION:")
			for section, fields := range themeAnalysis.FrontmatterFields {
				sb.WriteString(fmt.Sprintf("\n  - %s: %s", section, strings.Join(fields, ", ")))
			}
		}

		// Page params the theme uses
		if len(configAnalysis.PageParams) > 0 {
			sb.WriteString(fmt.Sprintf("\nPAGE PARAMS: %s", strings.Join(configAnalysis.PageParams, ", ")))
		}

		// Taxonomies
		if themeAnalysis.HasTaxonomies {
			sb.WriteString("\nTAXONOMIES: tags, categories supported")
		}
	}

	sb.WriteString("\n\nCreate the JSON plan now.")

	return sb.String()
}

// BuildSinglePageUserPrompt builds the user prompt for generating a single page.
// frontmatterFields are the recommended fields for this page's section (from theme analysis).
func BuildSinglePageUserPrompt(plan *SitePlan, page *PageSpec, frontmatterFields []string) string {
	var sb strings.Builder

	// Site context
	sb.WriteString(fmt.Sprintf(`SITE CONTEXT:
- Name: %s
- Type: %s
- Description: %s
- Audience: %s`,
		plan.SiteName, plan.SiteType, plan.Description, plan.Audience))

	if plan.Tone != "" {
		sb.WriteString(fmt.Sprintf("\n- Tone: %s", plan.Tone))
	}
	if plan.Theme != "" {
		sb.WriteString(fmt.Sprintf("\n- Theme: %s", plan.Theme))
	}

	// Page context
	sb.WriteString(fmt.Sprintf(`

PAGE TO GENERATE:
- Path: %s
- ID: %s
- Type: %s`,
		page.Path, page.ID, page.PageType))

	if page.Title != "" {
		sb.WriteString(fmt.Sprintf("\n- Title: %s", page.Title))
	}

	if page.Description != "" {
		sb.WriteString(fmt.Sprintf("\n\nPAGE REQUIREMENTS:\n%s", page.Description))
	}

	// Include planner-provided frontmatter (title, description, etc. from the AI plan)
	if len(page.Frontmatter) > 0 {
		sb.WriteString("\n\nPLANNED FRONTMATTER:")
		for key, value := range page.Frontmatter {
			sb.WriteString(fmt.Sprintf("\n- %s: %v", key, value))
		}
	}

	if len(page.Keywords) > 0 {
		sb.WriteString(fmt.Sprintf("\n\nKEYWORDS: %s", strings.Join(page.Keywords, ", ")))
	}

	if page.WordCount > 0 {
		sb.WriteString(fmt.Sprintf("\n\nTARGET WORDS: ~%d", page.WordCount))
	}

	if len(page.InternalLinks) > 0 {
		sb.WriteString(fmt.Sprintf("\n\nLINK TO: %s", strings.Join(page.InternalLinks, ", ")))
	}

	// Required frontmatter fields (from theme analysis, passed as parameter)
	if len(frontmatterFields) > 0 {
		sb.WriteString(fmt.Sprintf("\n\nREQUIRED FRONTMATTER FIELDS: %s", strings.Join(frontmatterFields, ", ")))
	}

	sb.WriteString("\n\nGenerate the complete Hugo markdown file. Start with ---")

	return sb.String()
}

// BuildUpdatePrompt builds the user prompt for content updates
func BuildUpdatePrompt(instruction, existingContent string) string {
	return fmt.Sprintf(`EXISTING CONTENT:
---START---
%s
---END---

UPDATE INSTRUCTIONS: %s

Return the complete updated Hugo markdown file.`, existingContent, instruction)
}

// BuildUserPrompt builds a user prompt for generic content generation
func BuildUserPrompt(instruction, context string) string {
	if context != "" {
		return fmt.Sprintf(`INSTRUCTION: %s

CONTEXT:
%s

Generate the complete Hugo content file.`, instruction, context)
	}
	return fmt.Sprintf(`INSTRUCTION: %s

Generate the complete Hugo content file.`, instruction)
}

// =============================================================================
// 7. UTILITIES
// =============================================================================

// CleanMarkdownFences removes markdown code fences from AI-generated content
func CleanMarkdownFences(content string) string {
	content = strings.TrimSpace(content)

	// Remove code fences (try specific first)
	prefixes := []string{"```markdown\n", "```md\n", "```yaml\n", "```\n"}
	for _, prefix := range prefixes {
		content = strings.TrimPrefix(content, prefix)
	}

	suffixes := []string{"\n```", "```"}
	for _, suffix := range suffixes {
		content = strings.TrimSuffix(content, suffix)
	}

	return strings.TrimSpace(content)
}

// CleanGeneratedContent cleans AI-generated Hugo content
func CleanGeneratedContent(content string) string {
	content = CleanMarkdownFences(content)

	// Ensure draft: false
	draftPatterns := []string{
		"draft: true", "draft:true",
		"draft: True", "draft:True",
		"draft: TRUE",
	}

	for _, pattern := range draftPatterns {
		if strings.Contains(content, pattern) {
			content = strings.Replace(content, pattern, "draft: false", 1)
			break
		}
	}

	return content
}

// =============================================================================
// ACTIVE EXPORTS
// =============================================================================

// SystemPromptSitePlanner is the system prompt for site planning
var SystemPromptSitePlanner = ComposeSitePlannerPrompt()

// SystemPromptContentUpdate is the system prompt for content updates
var SystemPromptContentUpdate = ComposeContentUpdatePrompt()

// =============================================================================
// SMART CONTENT GENERATOR (used by content_generator.go)
// =============================================================================

// BuildSmartContentPrompt builds the system prompt for the smart content generator (walgo ai new).
// It reuses the shared rule blocks and adds structure-aware context + response format.
func BuildSmartContentPrompt(structureInfo string) string {
	var sb strings.Builder

	sb.WriteString(`You are an expert Hugo content generator.

CONTENT STRUCTURE:
`)
	sb.WriteString(structureInfo)
	sb.WriteString("\n\n")
	sb.WriteString(OutputFormatRules)
	sb.WriteString("\n\n")
	sb.WriteString(YAMLSyntaxRules)
	sb.WriteString("\n\n")
	sb.WriteString(HugoRules)
	sb.WriteString("\n\n")
	sb.WriteString(ContentQualityRules)
	sb.WriteString("\n\n")
	sb.WriteString(SEORules)
	sb.WriteString(`

RESPONSE FORMAT:
CONTENT_TYPE: <folder_name>
FILENAME: <filename.md>
---
title: "..."
date: ISO8601
draft: false
---
<content>

Return ONLY the Hugo markdown file.`)

	return sb.String()
}

// BuildContentStructureInfo formats content structure information
func BuildContentStructureInfo(contentTypes []struct {
	Name      string
	FileCount int
	Files     []string
}) string {
	if len(contentTypes) == 0 {
		return "No existing content types found. Will create appropriate structure."
	}

	var sb strings.Builder
	sb.WriteString("Available content types:\n")
	for _, ct := range contentTypes {
		sb.WriteString(fmt.Sprintf("- %s/ (%d files)\n", ct.Name, ct.FileCount))
		if len(ct.Files) > 0 && len(ct.Files) <= 5 {
			for _, f := range ct.Files {
				sb.WriteString(fmt.Sprintf("  - %s\n", f))
			}
		}
	}
	return sb.String()
}
