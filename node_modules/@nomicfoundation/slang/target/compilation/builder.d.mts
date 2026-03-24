import { Cursor } from "../cst/index.mjs";
import { CompilationUnit } from "./index.mjs";
/**
 * User-provided options and callbacks necessary for the `CompilationBuilder` class to perform its job.
 */
export interface CompilationBuilderConfig {
    /**
     * The language version to parse files with.
     */
    languageVersion: string;
    /**
     * Callback used by this builder to load the contents of a file.
     *
     * The user is responsible for fetching the file from the filesystem.
     * If the file is not found, the callback should return undefined.
     * Any errors thrown by the callback will be propagated to the caller.
     */
    readFile: (fileId: string) => Promise<string | undefined>;
    /**
     * Callback used by this builder to resolve an import path.
     * For example, if a source file contains the following statement:
     *
     * ```solidity
     * import {Foo} from "foo.sol";
     * ```
     *
     * Then the API will invoke the callback with a cursor pointing to the `"foo.sol"` string literal.
     *
     * The user is responsible for resolving it to a file in the compilation, and return its ID.
     * If the callback returns `undefined`, the import will stay unresolved.
     * Any errors thrown by the callback will be propagated to the caller.
     */
    resolveImport: (sourceFileId: string, importPath: Cursor) => Promise<string | undefined>;
}
/**
 * A builder for creating compilation units.
 * Allows incrementally building a list of all files and their imports.
 */
export declare class CompilationBuilder {
    private readonly internalBuilder;
    /**
     * The user-supplied configuration.
     */
    readonly config: CompilationBuilderConfig;
    private readonly seenFiles;
    private constructor();
    /**
     * Creates a new compilation builder for the specified language version.
     */
    static create(config: CompilationBuilderConfig): CompilationBuilder;
    /**
     * Adds a source file to the compilation unit.
     * Typically, users only need to add the "root" file, which contains the main contract they are trying to analyze.
     * Any files that are imported by the root file will be discovered and loaded automatically by the config callbacks.
     *
     * Adding multiple files (roots) is supported. For example, an IDE can choose to add all NPM dependencies,
     * regardless of whether they are imported or not, to be able to query the definitions there.
     *
     * Adding a file that has already been added is a no-op.
     */
    addFile(id: string): Promise<void>;
    /**
     * Builds and returns the final compilation unit.
     */
    build(): CompilationUnit;
}
//# sourceMappingURL=builder.d.mts.map