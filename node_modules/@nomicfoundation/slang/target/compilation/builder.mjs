var __awaiter = (this && this.__awaiter) || function (thisArg, _arguments, P, generator) {
    function adopt(value) { return value instanceof P ? value : new P(function (resolve) { resolve(value); }); }
    return new (P || (P = Promise))(function (resolve, reject) {
        function fulfilled(value) { try { step(generator.next(value)); } catch (e) { reject(e); } }
        function rejected(value) { try { step(generator["throw"](value)); } catch (e) { reject(e); } }
        function step(result) { result.done ? resolve(result.value) : adopt(result.value).then(fulfilled, rejected); }
        step((generator = generator.apply(thisArg, _arguments || [])).next());
    });
};
import * as wasm from "../../wasm/index.mjs";
/**
 * A builder for creating compilation units.
 * Allows incrementally building a list of all files and their imports.
 */
export class CompilationBuilder {
    constructor(internalBuilder, 
    /**
     * The user-supplied configuration.
     */
    config) {
        this.internalBuilder = internalBuilder;
        this.config = config;
        this.seenFiles = new Set();
    }
    /**
     * Creates a new compilation builder for the specified language version.
     */
    static create(config) {
        const internalBuilder = wasm.compilation.InternalCompilationBuilder.create(config.languageVersion);
        return new CompilationBuilder(internalBuilder, config);
    }
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
    addFile(id) {
        return __awaiter(this, void 0, void 0, function* () {
            if (this.seenFiles.has(id)) {
                return;
            }
            else {
                this.seenFiles.add(id);
            }
            const contents = yield this.config.readFile(id);
            if (contents === undefined) {
                return;
            }
            const { importPaths } = this.internalBuilder.addFile(id, contents);
            yield Promise.all(importPaths.map((importPath) => __awaiter(this, void 0, void 0, function* () {
                const destinationFileId = yield this.config.resolveImport(id, importPath);
                if (destinationFileId === undefined) {
                    return;
                }
                this.internalBuilder.resolveImport(id, importPath, destinationFileId);
                yield this.addFile(destinationFileId);
            })));
        });
    }
    /**
     * Builds and returns the final compilation unit.
     */
    build() {
        return this.internalBuilder.build();
    }
}
//# sourceMappingURL=builder.mjs.map