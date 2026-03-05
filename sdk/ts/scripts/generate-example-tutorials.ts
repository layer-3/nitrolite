import * as fs from 'fs';
import * as path from 'path';
import { glob } from 'glob';

interface TutorialSection {
    type: 'markdown' | 'code' | 'architecture' | 'step' | 'concept';
    content: string;
    title?: string;
    file?: string;
    lineStart?: number;
    lineEnd?: number;
    language?: string;
    metadata?: Record<string, any>;
}

interface TutorialStructure {
    title: string;
    description: string;
    difficulty: 'beginner' | 'intermediate' | 'advanced';
    estimatedTime: string;
    prerequisites: string[];
    sections: TutorialSection[];
    metadata: {
        lastUpdated: string;
        examplePath: string;
        technologies: string[];
        concepts: string[];
    };
}

class ExampleTutorialGenerator {
    private examplesDir: string;
    private outputDir: string;

    constructor(examplesDir: string = '../../examples', outputDir: string = './docs/tutorials') {
        this.examplesDir = path.resolve(__dirname, examplesDir);
        this.outputDir = path.resolve(__dirname, '..', outputDir);
    }

    /**
     * Generate tutorials from all example projects
     */
    async generateAllTutorials(): Promise<void> {
        console.log('Generating tutorials from example projects...');

        // Ensure output directory exists
        if (!fs.existsSync(this.outputDir)) {
            fs.mkdirSync(this.outputDir, { recursive: true });
        }

        // Find all example project directories
        const exampleDirs = fs
            .readdirSync(this.examplesDir)
            .filter((dir) => fs.statSync(path.join(this.examplesDir, dir)).isDirectory());

        console.log(`Found ${exampleDirs.length} example projects: ${exampleDirs.join(', ')}`);

        for (const exampleDir of exampleDirs) {
            await this.generateTutorialForProject(exampleDir);
        }

        // Generate index file
        await this.generateTutorialIndex(exampleDirs);

        console.log('✅ Tutorial generation complete!');
    }

    /**
     * Generate tutorial for a specific example project
     */
    private async generateTutorialForProject(projectName: string): Promise<void> {
        console.log(`Processing ${projectName}...`);

        const projectPath = path.join(this.examplesDir, projectName);
        const tutorial = await this.parseTutorialFromProject(projectPath, projectName);

        if (!tutorial) {
            console.log(`⚠️  No tutorial content found in ${projectName}`);
            return;
        }

        const outputFile = path.join(this.outputDir, `${projectName}.md`);
        const markdown = this.generateMarkdownFromTutorial(tutorial);

        fs.writeFileSync(outputFile, markdown);
        console.log(`✅ Generated tutorial: ${outputFile}`);
    }

    /**
     * Parse tutorial content from TypeScript files in a project
     */
    private async parseTutorialFromProject(
        projectPath: string,
        projectName: string,
    ): Promise<TutorialStructure | null> {
        // Find all TypeScript files in the project
        const tsFiles = await glob('**/*.{ts,tsx}', {
            cwd: projectPath,
            ignore: ['node_modules/**', 'dist/**', 'build/**'],
        });

        if (tsFiles.length === 0) {
            return null;
        }

        const sections: TutorialSection[] = [];
        let tutorialMeta: any = {};

        // Parse each TypeScript file for documentation comments
        for (const file of tsFiles) {
            const filePath = path.join(projectPath, file);
            const content = fs.readFileSync(filePath, 'utf-8');

            const fileSections = this.parseFileForDocumentation(content, file);
            sections.push(...fileSections);

            // Look for tutorial metadata in special comments
            const meta = this.extractTutorialMetadata(content);
            if (meta) {
                tutorialMeta = { ...tutorialMeta, ...meta };
            }
        }

        // Parse README.md if it exists
        const readmePaths = [
            path.join(projectPath, 'README.md'),
            path.join(projectPath, 'Readme.md'),
            path.join(projectPath, 'readme.md'),
        ];

        for (const readmePath of readmePaths) {
            if (fs.existsSync(readmePath)) {
                const readmeContent = fs.readFileSync(readmePath, 'utf-8');
                const readmeSections = this.parseReadmeForDocumentation(readmeContent);
                sections.push(...readmeSections);
                break;
            }
        }

        if (sections.length === 0) {
            return null;
        }

        // Build tutorial structure
        return {
            title: tutorialMeta.title || `${projectName.charAt(0).toUpperCase() + projectName.slice(1)} Tutorial`,
            description: tutorialMeta.description || `Learn how to build ${projectName} with Nitrolite`,
            difficulty: tutorialMeta.difficulty || 'intermediate',
            estimatedTime: tutorialMeta.estimatedTime || '30-45 minutes',
            prerequisites: tutorialMeta.prerequisites || [
                'Basic TypeScript knowledge',
                'Understanding of state channels',
            ],
            sections: this.organizeSections(sections),
            metadata: {
                lastUpdated: new Date().toISOString().split('T')[0],
                examplePath: projectName,
                technologies: tutorialMeta.technologies || ['TypeScript', 'Nitrolite'],
                concepts: tutorialMeta.concepts || ['State Channels', 'Real-time Applications'],
            },
        };
    }

    /**
     * Parse a TypeScript file for documentation comments
     */
    private parseFileForDocumentation(content: string, filename: string): TutorialSection[] {
        const sections: TutorialSection[] = [];
        const lines = content.split('\n');

        let currentSection: Partial<TutorialSection> | null = null;
        let inDocBlock = false;
        let inCodeBlock = false;
        let blockStart = 0;

        for (let i = 0; i < lines.length; i++) {
            const line = lines[i];

            // Check for tutorial documentation blocks
            if (line.trim().startsWith('/**tutorial') || line.trim().startsWith('/*tutorial')) {
                inDocBlock = true;
                blockStart = i;
                currentSection = {
                    type: 'markdown',
                    content: '',
                    file: filename,
                    lineStart: i + 1,
                };

                // Extract section type and metadata from the opening comment
                const match = line.match(/\/\*\*?tutorial:(\w+)(?:\s+(.+))?/);
                if (match) {
                    currentSection.type = match[1] as any;
                    if (match[2]) {
                        try {
                            currentSection.metadata = JSON.parse(match[2]);
                        } catch {
                            currentSection.title = match[2];
                        }
                    }
                }
                continue;
            }

            // Check for end of documentation block
            if (inDocBlock && line.trim().endsWith('*/')) {
                inDocBlock = false;
                if (currentSection) {
                    currentSection.lineEnd = i + 1;
                    sections.push(currentSection as TutorialSection);
                    currentSection = null;
                }
                continue;
            }

            // Process documentation content
            if (inDocBlock && currentSection) {
                let docLine = line.replace(/^\s*\*\s?/, '');
                currentSection.content += docLine + '\n';
            }

            // Capture code blocks that follow documentation
            if (!inDocBlock && sections.length > 0) {
                const lastSection = sections[sections.length - 1];
                if (lastSection.type === 'step' && !lastSection.metadata?.codeExtracted) {
                    // Extract the next meaningful code block
                    if (line.trim() && !line.startsWith('//') && !line.startsWith('/*')) {
                        const codeBlock = this.extractCodeBlock(lines, i, filename);
                        if (codeBlock) {
                            sections.push({
                                type: 'code',
                                content: codeBlock.content,
                                file: filename,
                                lineStart: codeBlock.start,
                                lineEnd: codeBlock.end,
                                language: this.getLanguageFromFile(filename),
                            });
                            lastSection.metadata = {
                                ...(lastSection.metadata || {}),
                                codeExtracted: true,
                            };
                        }
                    }
                }
            }
        }

        return sections;
    }

    /**
     * Extract tutorial metadata from special comments
     */
    private extractTutorialMetadata(content: string): any {
        const metaMatch = content.match(/\/\*\*tutorial-meta\s*([\s\S]*?)\*\//);
        if (!metaMatch) return null;

        try {
            return JSON.parse(metaMatch[1]);
        } catch {
            // Parse YAML-like format
            const meta: any = {};
            const lines = metaMatch[1].split('\n');
            for (const line of lines) {
                const match = line.match(/^\s*(\w+):\s*(.+)$/);
                if (match) {
                    const key = match[1];
                    let value: any = match[2].trim();

                    // Parse arrays
                    if (value.startsWith('[') && value.endsWith(']')) {
                        value = value
                            .slice(1, -1)
                            .split(',')
                            .map((s: string) => s.trim().replace(/['"]/g, ''));
                    }
                    // Parse strings
                    else if (value.startsWith('"') && value.endsWith('"')) {
                        value = value.slice(1, -1);
                    }

                    meta[key] = value;
                }
            }
            return meta;
        }
    }

    /**
     * Parse README.md for additional documentation
     */
    private parseReadmeForDocumentation(content: string): TutorialSection[] {
        const sections: TutorialSection[] = [];

        // Extract main description and features
        const lines = content.split('\n');
        let currentSection = '';
        let inFeatures = false;

        for (const line of lines) {
            if (line.startsWith('# ')) {
                continue; // Skip title
            }
            if (line.startsWith('## Features')) {
                inFeatures = true;
                continue;
            }
            if (line.startsWith('## ') && inFeatures) {
                if (currentSection) {
                    sections.push({
                        type: 'concept',
                        title: 'Project Features',
                        content: currentSection,
                    });
                }
                break;
            }
            if (inFeatures) {
                currentSection += line + '\n';
            }
        }

        return sections;
    }

    /**
     * Extract a code block starting from a given line
     */
    private extractCodeBlock(
        lines: string[],
        startIndex: number,
        filename: string,
    ): { content: string; start: number; end: number } | null {
        let braceCount = 0;
        let parenCount = 0;
        let inFunction = false;
        let blockStart = startIndex;
        let blockEnd = startIndex;

        // Find the start of a meaningful code construct
        for (let i = startIndex; i < lines.length && i < startIndex + 20; i++) {
            const line = lines[i].trim();
            if (!line || line.startsWith('//') || line.startsWith('/*')) continue;

            // Look for function declarations, class methods, const declarations, etc.
            if (
                line.match(/^(export\s+)?(const|function|class|interface)\s+/) ||
                line.match(/^\w+\s*[:=]\s*\(/) ||
                line.includes('=>')
            ) {
                blockStart = i;
                inFunction = true;
                break;
            }
        }

        if (!inFunction) return null;

        // Extract the complete block
        for (let i = blockStart; i < lines.length; i++) {
            const line = lines[i];
            blockEnd = i;

            // Count braces and parentheses
            for (const char of line) {
                if (char === '{') braceCount++;
                if (char === '}') braceCount--;
                if (char === '(') parenCount++;
                if (char === ')') parenCount--;
            }

            // Stop when we've closed all braces and have a complete statement
            if (
                braceCount === 0 &&
                parenCount === 0 &&
                (line.includes(';') || line.includes('}') || i - blockStart > 10)
            ) {
                break;
            }

            // Safety limit
            if (i - blockStart > 50) break;
        }

        const content = lines
            .slice(blockStart, blockEnd + 1)
            .map((line) => line.replace(/^\s{0,2}/, '')) // Remove minimal indentation
            .join('\n');

        return {
            content,
            start: blockStart + 1,
            end: blockEnd + 1,
        };
    }

    /**
     * Organize sections into a logical flow
     */
    private organizeSections(sections: TutorialSection[]): TutorialSection[] {
        // Define section order: architecture first, then concepts, then steps, then code
        const order = ['architecture', 'concept', 'step', 'code', 'markdown'];

        return sections.sort((a, b) => {
            const aIndex = order.indexOf(a.type);
            const bIndex = order.indexOf(b.type);

            if (aIndex !== bIndex) {
                return aIndex - bIndex;
            }

            // Within the same type, maintain original order
            return 0;
        });
    }

    /**
     * Generate a unique anchor ID from a title
     */
    private generateAnchorId(title: string): string {
        return title
            .toLowerCase()
            .replace(/[^a-z0-9\s-]/g, '') // Remove special characters
            .replace(/\s+/g, '-') // Replace spaces with hyphens
            .replace(/-+/g, '-') // Replace multiple hyphens with single
            .replace(/^-|-$/g, ''); // Remove leading/trailing hyphens
    }

    /**
     * Generate markdown from tutorial structure
     */
    private generateMarkdownFromTutorial(tutorial: TutorialStructure): string {
        let markdown = '';

        // Header
        markdown += `# ${tutorial.title}\n\n`;
        markdown += `> ${tutorial.description}\n\n`;

        // Metadata
        markdown += `## Tutorial Info\n\n`;
        markdown += `| Property | Value |\n`;
        markdown += `|----------|-------|\n`;
        markdown += `| **Difficulty** | ${tutorial.difficulty} |\n`;
        markdown += `| **Estimated Time** | ${tutorial.estimatedTime} |\n`;
        markdown += `| **Last Updated** | ${tutorial.metadata.lastUpdated} |\n`;
        markdown += `| **Technologies** | ${tutorial.metadata.technologies.join(', ')} |\n\n`;

        // Prerequisites
        if (tutorial.prerequisites.length > 0) {
            markdown += `## Prerequisites\n\n`;
            tutorial.prerequisites.forEach((prereq) => {
                markdown += `- ${prereq}\n`;
            });
            markdown += '\n';
        }

        // Key Concepts
        if (tutorial.metadata.concepts.length > 0) {
            markdown += `## What You'll Learn\n\n`;
            tutorial.metadata.concepts.forEach((concept) => {
                markdown += `- ${concept}\n`;
            });
            markdown += '\n';
        }

        // Collect sections for table of contents
        const architectureSections = tutorial.sections.filter((s) => s.type === 'architecture');
        const conceptSections = tutorial.sections.filter((s) => s.type === 'concept');
        const stepSections = tutorial.sections.filter((s) => s.type === 'step');

        // Table of Contents
        markdown += `## Table of Contents\n\n`;

        // Architecture sections
        architectureSections.forEach((section) => {
            const title = section.title || 'Architecture Overview';
            const anchorId = this.generateAnchorId(title);
            markdown += `- [${title}](#${anchorId})\n`;
        });

        // Concept sections
        conceptSections.forEach((section) => {
            const title = section.title || 'Key Concepts';
            const anchorId = this.generateAnchorId(title);
            markdown += `- [${title}](#${anchorId})\n`;
        });

        // Step sections
        stepSections.forEach((section, index) => {
            const stepNumber = index + 1;
            const title = section.title || `Step ${stepNumber}`;
            const anchorId = this.generateAnchorId(`step ${stepNumber} ${title}`);
            markdown += `${stepNumber}. [${title}](#${anchorId})\n`;
        });

        markdown += '\n';

        // Content sections
        let stepCount = 1;
        for (const section of tutorial.sections) {
            markdown += this.renderSection(section, stepCount);
            if (section.type === 'step') {
                stepCount++;
            }
        }

        // Footer
        markdown += `---\n\n`;
        markdown += `## Next Steps\n\n`;
        markdown += `Congratulations! You've successfully built a ${tutorial.title.toLowerCase()}.\n\n`;
        markdown += `### Continue Learning\n\n`;
        markdown += `- Explore the [SDK documentation](../README.md)\n`;
        markdown += `- Check out more [Examples](../../examples/)\n`;
        markdown += `- Report issues or contribute at [GitHub](https://github.com/layer-3/nitrolite)\n\n`;
        markdown += `### Improve This Tutorial\n\n`;
        markdown += `Found an issue or want to improve this tutorial? [Edit on GitHub](https://github.com/layer-3/nitrolite/tree/main/examples/${tutorial.metadata.examplePath})\n\n`;

        return markdown;
    }

    /**
     * Render a specific section type
     */
    private renderSection(section: TutorialSection, stepCount?: number): string {
        let markdown = '';

        switch (section.type) {
            case 'step':
                const stepTitle = section.title || 'Implementation';
                const stepAnchorId = this.generateAnchorId(`step ${stepCount} ${stepTitle}`);
                markdown += `## Step ${stepCount}: ${stepTitle} {#${stepAnchorId}}\n\n`;
                markdown += section.content + '\n\n';
                break;

            case 'code':
                const language = section.language || 'typescript';
                markdown += `\`\`\`${language}\n`;
                markdown += section.content;
                markdown += `\n\`\`\`\n\n`;
                if (section.file) {
                    markdown += `> **File:** \`${section.file}\` (lines ${section.lineStart}-${section.lineEnd})\n\n`;
                }
                break;

            case 'architecture':
                const archTitle = section.title || 'Architecture Overview';
                const archAnchorId = this.generateAnchorId(archTitle);
                markdown += `## ${archTitle} {#${archAnchorId}}\n\n`;
                markdown += section.content + '\n\n';
                break;

            case 'concept':
                const conceptTitle = section.title || 'Key Concepts';
                const conceptAnchorId = this.generateAnchorId(conceptTitle);
                markdown += `## ${conceptTitle} {#${conceptAnchorId}}\n\n`;
                markdown += section.content + '\n\n';
                break;

            case 'markdown':
            default:
                markdown += section.content + '\n\n';
                break;
        }

        return markdown;
    }

    /**
     * Get programming language from filename
     */
    private getLanguageFromFile(filename: string): string {
        const ext = path.extname(filename);
        const langMap: Record<string, string> = {
            '.ts': 'typescript',
            '.tsx': 'typescript',
            '.js': 'javascript',
            '.jsx': 'javascript',
            '.vue': 'vue',
            '.py': 'python',
            '.go': 'go',
            '.rs': 'rust',
        };
        return langMap[ext] || 'text';
    }

    /**
     * Generate index file for all tutorials
     */
    private async generateTutorialIndex(exampleDirs: string[]): Promise<void> {
        let markdown = '';

        markdown += `# Nitrolite Example Tutorials\n\n`;
        markdown += `Learn Nitrolite through hands-on examples and comprehensive tutorials.\n\n`;

        markdown += `## Available Tutorials\n\n`;

        for (const dir of exampleDirs) {
            const tutorialFile = path.join(this.outputDir, `${dir}.md`);
            if (fs.existsSync(tutorialFile)) {
                // Read the tutorial file to extract metadata
                const content = fs.readFileSync(tutorialFile, 'utf-8');
                const titleMatch = content.match(/^# (.+)$/m);
                const descMatch = content.match(/^> (.+)$/m);
                const difficultyMatch = content.match(/\| \*\*Difficulty\*\* \| (\w+) \|/);
                const timeMatch = content.match(/\| \*\*Estimated Time\*\* \| ([^|]+) \|/);

                const title = titleMatch ? titleMatch[1] : dir;
                const description = descMatch ? descMatch[1] : `Learn ${dir}`;
                const difficulty = difficultyMatch ? difficultyMatch[1] : 'intermediate';
                const time = timeMatch ? timeMatch[1].trim() : '30 minutes';

                markdown += `### [${title}](${dir}.md)\n\n`;
                markdown += `${description}\n\n`;
                markdown += `**Difficulty:** ${difficulty} | **Time:** ${time}\n\n`;
            }
        }

        markdown += `## Getting Started\n\n`;
        markdown += `1. Choose a tutorial based on your interest and experience level\n`;
        markdown += `2. Make sure you have the [prerequisites](../README.md#prerequisites) installed\n`;
        markdown += `3. Follow the step-by-step instructions\n`;
        markdown += `4. Experiment and build upon the examples\n\n`;

        markdown += `## Tutorial Format\n\n`;
        markdown += `Each tutorial follows a consistent structure:\n\n`;
        markdown += `- **Tutorial Info** - Difficulty, time estimate, and technologies\n`;
        markdown += `- **Prerequisites** - What you need to know beforehand\n`;
        markdown += `- **What You'll Learn** - Key concepts covered\n`;
        markdown += `- **Step-by-step Guide** - Detailed implementation walkthrough\n`;
        markdown += `- **Architecture Diagrams** - Visual explanations of concepts\n`;
        markdown += `- **Working Code** - Real examples with line references\n`;
        markdown += `- **Next Steps** - How to continue learning\n\n`;

        markdown += `## Contributing\n\n`;
        markdown += `Want to create a new tutorial? Here's how:\n\n`;
        markdown += `1. Create an example project in the \`examples/\` directory\n`;
        markdown += `2. Add tutorial documentation using special comments in your TypeScript files:\n\n`;
        markdown += `\`\`\`typescript\n`;
        markdown += `/**tutorial-meta\n`;
        markdown += `title: "My Amazing Tutorial"\n`;
        markdown += `description: "Learn how to build something amazing"\n`;
        markdown += `difficulty: "beginner"\n`;
        markdown += `estimatedTime: "20 minutes"\n`;
        markdown += `technologies: ["TypeScript", "React", "Nitrolite"]\n`;
        markdown += `concepts: ["State Channels", "Real-time Updates"]\n`;
        markdown += `*/\n\n`;
        markdown += `/**tutorial:step Setup the Project\n`;
        markdown += `First, we need to set up our project structure.\n`;
        markdown += `This involves creating the basic files and installing dependencies.\n`;
        markdown += `*/\n\n`;
        markdown += `/**tutorial:architecture\n`;
        markdown += `## System Architecture\n\n`;
        markdown += `Our application follows this structure:\n\n`;
        markdown += `\`\`\`\n`;
        markdown += `[Client] <-> [WebSocket] <-> [Server] <-> [Nitrolite]\n`;
        markdown += `\`\`\`\n`;
        markdown += `*/\n`;
        markdown += `\`\`\`\n\n`;
        markdown += `3. Run \`npm run docs:tutorials\` to generate the tutorial\n\n`;

        const indexPath = path.join(this.outputDir, 'index.md');
        fs.writeFileSync(indexPath, markdown);
        console.log(`✅ Generated tutorial index: ${indexPath}`);
    }
}

// Main execution
async function main() {
    const generator = new ExampleTutorialGenerator();
    await generator.generateAllTutorials();
}

if (require.main === module) {
    main().catch(console.error);
}

export { ExampleTutorialGenerator };
