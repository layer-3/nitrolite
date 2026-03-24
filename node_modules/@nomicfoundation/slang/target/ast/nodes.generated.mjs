// This file is generated automatically by infrastructure scripts. Please don't edit by hand.
import * as wasm from "../../wasm/index.mjs";
import { NonterminalKind } from "../cst/index.mjs";
//
// Sequences:
//
/**
 * This node represents a `SourceUnit` nonterminal, with the following structure:
 *
 * ```ebnf
 * SourceUnit = (* members: *) SourceUnitMembers;
 * ```
 */
export class SourceUnit {
    /**
     * Constructs a new AST node of type `SourceUnit`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `SourceUnit`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const [$members] = wasm.ast.Selectors.sequence(this.cst);
            return {
                members: new SourceUnitMembers($members),
            };
        });
        assertKind(this.cst.kind, NonterminalKind.SourceUnit);
    }
    /**
     * Returns the child node that has the label `members`.
     */
    get members() {
        return this.fetch().members;
    }
}
/**
 * This node represents a `PragmaDirective` nonterminal, with the following structure:
 *
 * ```ebnf
 * PragmaDirective = (* pragma_keyword: *) PRAGMA_KEYWORD
 *                   (* pragma: *) Pragma
 *                   (* semicolon: *) SEMICOLON;
 * ```
 */
export class PragmaDirective {
    /**
     * Constructs a new AST node of type `PragmaDirective`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `PragmaDirective`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const [$pragmaKeyword, $pragma, $semicolon] = wasm.ast.Selectors.sequence(this.cst);
            return {
                pragmaKeyword: $pragmaKeyword,
                pragma: new Pragma($pragma),
                semicolon: $semicolon,
            };
        });
        assertKind(this.cst.kind, NonterminalKind.PragmaDirective);
    }
    /**
     * Returns the child node that has the label `pragma_keyword`.
     */
    get pragmaKeyword() {
        return this.fetch().pragmaKeyword;
    }
    /**
     * Returns the child node that has the label `pragma`.
     */
    get pragma() {
        return this.fetch().pragma;
    }
    /**
     * Returns the child node that has the label `semicolon`.
     */
    get semicolon() {
        return this.fetch().semicolon;
    }
}
/**
 * This node represents a `AbicoderPragma` nonterminal, with the following structure:
 *
 * ```ebnf
 * (* Introduced in 0.7.5 *)
 * AbicoderPragma = (* abicoder_keyword: *) ABICODER_KEYWORD
 *                  (* version: *) AbicoderVersion;
 * ```
 */
export class AbicoderPragma {
    /**
     * Constructs a new AST node of type `AbicoderPragma`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `AbicoderPragma`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const [$abicoderKeyword, $version] = wasm.ast.Selectors.sequence(this.cst);
            return {
                abicoderKeyword: $abicoderKeyword,
                version: new AbicoderVersion($version),
            };
        });
        assertKind(this.cst.kind, NonterminalKind.AbicoderPragma);
    }
    /**
     * Returns the child node that has the label `abicoder_keyword`.
     */
    get abicoderKeyword() {
        return this.fetch().abicoderKeyword;
    }
    /**
     * Returns the child node that has the label `version`.
     */
    get version() {
        return this.fetch().version;
    }
}
/**
 * This node represents a `ExperimentalPragma` nonterminal, with the following structure:
 *
 * ```ebnf
 * (* Introduced in 0.4.16 *)
 * ExperimentalPragma = (* experimental_keyword: *) EXPERIMENTAL_KEYWORD
 *                      (* feature: *) ExperimentalFeature;
 * ```
 */
export class ExperimentalPragma {
    /**
     * Constructs a new AST node of type `ExperimentalPragma`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `ExperimentalPragma`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const [$experimentalKeyword, $feature] = wasm.ast.Selectors.sequence(this.cst);
            return {
                experimentalKeyword: $experimentalKeyword,
                feature: new ExperimentalFeature($feature),
            };
        });
        assertKind(this.cst.kind, NonterminalKind.ExperimentalPragma);
    }
    /**
     * Returns the child node that has the label `experimental_keyword`.
     */
    get experimentalKeyword() {
        return this.fetch().experimentalKeyword;
    }
    /**
     * Returns the child node that has the label `feature`.
     */
    get feature() {
        return this.fetch().feature;
    }
}
/**
 * This node represents a `VersionPragma` nonterminal, with the following structure:
 *
 * ```ebnf
 * VersionPragma = (* solidity_keyword: *) SOLIDITY_KEYWORD
 *                 (* sets: *) VersionExpressionSets;
 * ```
 */
export class VersionPragma {
    /**
     * Constructs a new AST node of type `VersionPragma`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `VersionPragma`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const [$solidityKeyword, $sets] = wasm.ast.Selectors.sequence(this.cst);
            return {
                solidityKeyword: $solidityKeyword,
                sets: new VersionExpressionSets($sets),
            };
        });
        assertKind(this.cst.kind, NonterminalKind.VersionPragma);
    }
    /**
     * Returns the child node that has the label `solidity_keyword`.
     */
    get solidityKeyword() {
        return this.fetch().solidityKeyword;
    }
    /**
     * Returns the child node that has the label `sets`.
     */
    get sets() {
        return this.fetch().sets;
    }
}
/**
 * This node represents a `VersionRange` nonterminal, with the following structure:
 *
 * ```ebnf
 * VersionRange = (* start: *) VersionLiteral
 *                (* minus: *) MINUS
 *                (* end: *) VersionLiteral;
 * ```
 */
export class VersionRange {
    /**
     * Constructs a new AST node of type `VersionRange`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `VersionRange`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const [$start, $minus, $end] = wasm.ast.Selectors.sequence(this.cst);
            return {
                start: new VersionLiteral($start),
                minus: $minus,
                end: new VersionLiteral($end),
            };
        });
        assertKind(this.cst.kind, NonterminalKind.VersionRange);
    }
    /**
     * Returns the child node that has the label `start`.
     */
    get start() {
        return this.fetch().start;
    }
    /**
     * Returns the child node that has the label `minus`.
     */
    get minus() {
        return this.fetch().minus;
    }
    /**
     * Returns the child node that has the label `end`.
     */
    get end() {
        return this.fetch().end;
    }
}
/**
 * This node represents a `VersionTerm` nonterminal, with the following structure:
 *
 * ```ebnf
 * VersionTerm = (* operator: *) VersionOperator?
 *               (* literal: *) VersionLiteral;
 * ```
 */
export class VersionTerm {
    /**
     * Constructs a new AST node of type `VersionTerm`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `VersionTerm`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const [$operator, $literal] = wasm.ast.Selectors.sequence(this.cst);
            return {
                operator: $operator === undefined ? undefined : new VersionOperator($operator),
                literal: new VersionLiteral($literal),
            };
        });
        assertKind(this.cst.kind, NonterminalKind.VersionTerm);
    }
    /**
     * Returns the child node that has the label `operator`.
     */
    get operator() {
        return this.fetch().operator;
    }
    /**
     * Returns the child node that has the label `literal`.
     */
    get literal() {
        return this.fetch().literal;
    }
}
/**
 * This node represents a `ImportDirective` nonterminal, with the following structure:
 *
 * ```ebnf
 * ImportDirective = (* import_keyword: *) IMPORT_KEYWORD
 *                   (* clause: *) ImportClause
 *                   (* semicolon: *) SEMICOLON;
 * ```
 */
export class ImportDirective {
    /**
     * Constructs a new AST node of type `ImportDirective`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `ImportDirective`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const [$importKeyword, $clause, $semicolon] = wasm.ast.Selectors.sequence(this.cst);
            return {
                importKeyword: $importKeyword,
                clause: new ImportClause($clause),
                semicolon: $semicolon,
            };
        });
        assertKind(this.cst.kind, NonterminalKind.ImportDirective);
    }
    /**
     * Returns the child node that has the label `import_keyword`.
     */
    get importKeyword() {
        return this.fetch().importKeyword;
    }
    /**
     * Returns the child node that has the label `clause`.
     */
    get clause() {
        return this.fetch().clause;
    }
    /**
     * Returns the child node that has the label `semicolon`.
     */
    get semicolon() {
        return this.fetch().semicolon;
    }
}
/**
 * This node represents a `PathImport` nonterminal, with the following structure:
 *
 * ```ebnf
 * PathImport = (* path: *) StringLiteral
 *              (* alias: *) ImportAlias?;
 * ```
 */
export class PathImport {
    /**
     * Constructs a new AST node of type `PathImport`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `PathImport`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const [$path, $alias] = wasm.ast.Selectors.sequence(this.cst);
            return {
                path: new StringLiteral($path),
                alias: $alias === undefined ? undefined : new ImportAlias($alias),
            };
        });
        assertKind(this.cst.kind, NonterminalKind.PathImport);
    }
    /**
     * Returns the child node that has the label `path`.
     */
    get path() {
        return this.fetch().path;
    }
    /**
     * Returns the child node that has the label `alias`.
     */
    get alias() {
        return this.fetch().alias;
    }
}
/**
 * This node represents a `NamedImport` nonterminal, with the following structure:
 *
 * ```ebnf
 * NamedImport = (* asterisk: *) ASTERISK
 *               (* alias: *) ImportAlias
 *               (* from_keyword: *) FROM_KEYWORD
 *               (* path: *) StringLiteral;
 * ```
 */
export class NamedImport {
    /**
     * Constructs a new AST node of type `NamedImport`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `NamedImport`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const [$asterisk, $alias, $fromKeyword, $path] = wasm.ast.Selectors.sequence(this.cst);
            return {
                asterisk: $asterisk,
                alias: new ImportAlias($alias),
                fromKeyword: $fromKeyword,
                path: new StringLiteral($path),
            };
        });
        assertKind(this.cst.kind, NonterminalKind.NamedImport);
    }
    /**
     * Returns the child node that has the label `asterisk`.
     */
    get asterisk() {
        return this.fetch().asterisk;
    }
    /**
     * Returns the child node that has the label `alias`.
     */
    get alias() {
        return this.fetch().alias;
    }
    /**
     * Returns the child node that has the label `from_keyword`.
     */
    get fromKeyword() {
        return this.fetch().fromKeyword;
    }
    /**
     * Returns the child node that has the label `path`.
     */
    get path() {
        return this.fetch().path;
    }
}
/**
 * This node represents a `ImportDeconstruction` nonterminal, with the following structure:
 *
 * ```ebnf
 * ImportDeconstruction = (* open_brace: *) OPEN_BRACE
 *                        (* symbols: *) ImportDeconstructionSymbols
 *                        (* close_brace: *) CLOSE_BRACE
 *                        (* from_keyword: *) FROM_KEYWORD
 *                        (* path: *) StringLiteral;
 * ```
 */
export class ImportDeconstruction {
    /**
     * Constructs a new AST node of type `ImportDeconstruction`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `ImportDeconstruction`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const [$openBrace, $symbols, $closeBrace, $fromKeyword, $path] = wasm.ast.Selectors.sequence(this.cst);
            return {
                openBrace: $openBrace,
                symbols: new ImportDeconstructionSymbols($symbols),
                closeBrace: $closeBrace,
                fromKeyword: $fromKeyword,
                path: new StringLiteral($path),
            };
        });
        assertKind(this.cst.kind, NonterminalKind.ImportDeconstruction);
    }
    /**
     * Returns the child node that has the label `open_brace`.
     */
    get openBrace() {
        return this.fetch().openBrace;
    }
    /**
     * Returns the child node that has the label `symbols`.
     */
    get symbols() {
        return this.fetch().symbols;
    }
    /**
     * Returns the child node that has the label `close_brace`.
     */
    get closeBrace() {
        return this.fetch().closeBrace;
    }
    /**
     * Returns the child node that has the label `from_keyword`.
     */
    get fromKeyword() {
        return this.fetch().fromKeyword;
    }
    /**
     * Returns the child node that has the label `path`.
     */
    get path() {
        return this.fetch().path;
    }
}
/**
 * This node represents a `ImportDeconstructionSymbol` nonterminal, with the following structure:
 *
 * ```ebnf
 * ImportDeconstructionSymbol = (* name: *) IDENTIFIER
 *                              (* alias: *) ImportAlias?;
 * ```
 */
export class ImportDeconstructionSymbol {
    /**
     * Constructs a new AST node of type `ImportDeconstructionSymbol`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `ImportDeconstructionSymbol`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const [$name, $alias] = wasm.ast.Selectors.sequence(this.cst);
            return {
                name: $name,
                alias: $alias === undefined ? undefined : new ImportAlias($alias),
            };
        });
        assertKind(this.cst.kind, NonterminalKind.ImportDeconstructionSymbol);
    }
    /**
     * Returns the child node that has the label `name`.
     */
    get name() {
        return this.fetch().name;
    }
    /**
     * Returns the child node that has the label `alias`.
     */
    get alias() {
        return this.fetch().alias;
    }
}
/**
 * This node represents a `ImportAlias` nonterminal, with the following structure:
 *
 * ```ebnf
 * ImportAlias = (* as_keyword: *) AS_KEYWORD
 *               (* identifier: *) IDENTIFIER;
 * ```
 */
export class ImportAlias {
    /**
     * Constructs a new AST node of type `ImportAlias`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `ImportAlias`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const [$asKeyword, $identifier] = wasm.ast.Selectors.sequence(this.cst);
            return {
                asKeyword: $asKeyword,
                identifier: $identifier,
            };
        });
        assertKind(this.cst.kind, NonterminalKind.ImportAlias);
    }
    /**
     * Returns the child node that has the label `as_keyword`.
     */
    get asKeyword() {
        return this.fetch().asKeyword;
    }
    /**
     * Returns the child node that has the label `identifier`.
     */
    get identifier() {
        return this.fetch().identifier;
    }
}
/**
 * This node represents a `UsingDirective` nonterminal, with the following structure:
 *
 * ```ebnf
 * UsingDirective = (* using_keyword: *) USING_KEYWORD
 *                  (* clause: *) UsingClause
 *                  (* for_keyword: *) FOR_KEYWORD
 *                  (* target: *) UsingTarget
 *                  (* global_keyword: *) GLOBAL_KEYWORD? (* Introduced in 0.8.13 *)
 *                  (* semicolon: *) SEMICOLON;
 * ```
 */
export class UsingDirective {
    /**
     * Constructs a new AST node of type `UsingDirective`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `UsingDirective`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const [$usingKeyword, $clause, $forKeyword, $target, $globalKeyword, $semicolon] = wasm.ast.Selectors.sequence(this.cst);
            return {
                usingKeyword: $usingKeyword,
                clause: new UsingClause($clause),
                forKeyword: $forKeyword,
                target: new UsingTarget($target),
                globalKeyword: $globalKeyword === undefined ? undefined : $globalKeyword,
                semicolon: $semicolon,
            };
        });
        assertKind(this.cst.kind, NonterminalKind.UsingDirective);
    }
    /**
     * Returns the child node that has the label `using_keyword`.
     */
    get usingKeyword() {
        return this.fetch().usingKeyword;
    }
    /**
     * Returns the child node that has the label `clause`.
     */
    get clause() {
        return this.fetch().clause;
    }
    /**
     * Returns the child node that has the label `for_keyword`.
     */
    get forKeyword() {
        return this.fetch().forKeyword;
    }
    /**
     * Returns the child node that has the label `target`.
     */
    get target() {
        return this.fetch().target;
    }
    /**
     * Returns the child node that has the label `global_keyword`.
     */
    get globalKeyword() {
        return this.fetch().globalKeyword;
    }
    /**
     * Returns the child node that has the label `semicolon`.
     */
    get semicolon() {
        return this.fetch().semicolon;
    }
}
/**
 * This node represents a `UsingDeconstruction` nonterminal, with the following structure:
 *
 * ```ebnf
 * (* Introduced in 0.8.13 *)
 * UsingDeconstruction = (* open_brace: *) OPEN_BRACE
 *                       (* symbols: *) UsingDeconstructionSymbols
 *                       (* close_brace: *) CLOSE_BRACE;
 * ```
 */
export class UsingDeconstruction {
    /**
     * Constructs a new AST node of type `UsingDeconstruction`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `UsingDeconstruction`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const [$openBrace, $symbols, $closeBrace] = wasm.ast.Selectors.sequence(this.cst);
            return {
                openBrace: $openBrace,
                symbols: new UsingDeconstructionSymbols($symbols),
                closeBrace: $closeBrace,
            };
        });
        assertKind(this.cst.kind, NonterminalKind.UsingDeconstruction);
    }
    /**
     * Returns the child node that has the label `open_brace`.
     */
    get openBrace() {
        return this.fetch().openBrace;
    }
    /**
     * Returns the child node that has the label `symbols`.
     */
    get symbols() {
        return this.fetch().symbols;
    }
    /**
     * Returns the child node that has the label `close_brace`.
     */
    get closeBrace() {
        return this.fetch().closeBrace;
    }
}
/**
 * This node represents a `UsingDeconstructionSymbol` nonterminal, with the following structure:
 *
 * ```ebnf
 * (* Introduced in 0.8.13 *)
 * UsingDeconstructionSymbol = (* name: *) IdentifierPath
 *                             (* alias: *) UsingAlias?; (* Introduced in 0.8.19 *)
 * ```
 */
export class UsingDeconstructionSymbol {
    /**
     * Constructs a new AST node of type `UsingDeconstructionSymbol`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `UsingDeconstructionSymbol`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const [$name, $alias] = wasm.ast.Selectors.sequence(this.cst);
            return {
                name: new IdentifierPath($name),
                alias: $alias === undefined ? undefined : new UsingAlias($alias),
            };
        });
        assertKind(this.cst.kind, NonterminalKind.UsingDeconstructionSymbol);
    }
    /**
     * Returns the child node that has the label `name`.
     */
    get name() {
        return this.fetch().name;
    }
    /**
     * Returns the child node that has the label `alias`.
     */
    get alias() {
        return this.fetch().alias;
    }
}
/**
 * This node represents a `UsingAlias` nonterminal, with the following structure:
 *
 * ```ebnf
 * (* Introduced in 0.8.19 *)
 * UsingAlias = (* as_keyword: *) AS_KEYWORD
 *              (* operator: *) UsingOperator;
 * ```
 */
export class UsingAlias {
    /**
     * Constructs a new AST node of type `UsingAlias`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `UsingAlias`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const [$asKeyword, $operator] = wasm.ast.Selectors.sequence(this.cst);
            return {
                asKeyword: $asKeyword,
                operator: new UsingOperator($operator),
            };
        });
        assertKind(this.cst.kind, NonterminalKind.UsingAlias);
    }
    /**
     * Returns the child node that has the label `as_keyword`.
     */
    get asKeyword() {
        return this.fetch().asKeyword;
    }
    /**
     * Returns the child node that has the label `operator`.
     */
    get operator() {
        return this.fetch().operator;
    }
}
/**
 * This node represents a `ContractDefinition` nonterminal, with the following structure:
 *
 * ```ebnf
 * ContractDefinition = (* abstract_keyword: *) ABSTRACT_KEYWORD? (* Introduced in 0.6.0 *)
 *                      (* contract_keyword: *) CONTRACT_KEYWORD
 *                      (* name: *) IDENTIFIER
 *                      (* specifiers: *) ContractSpecifiers
 *                      (* open_brace: *) OPEN_BRACE
 *                      (* members: *) ContractMembers
 *                      (* close_brace: *) CLOSE_BRACE;
 * ```
 */
export class ContractDefinition {
    /**
     * Constructs a new AST node of type `ContractDefinition`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `ContractDefinition`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const [$abstractKeyword, $contractKeyword, $name, $specifiers, $openBrace, $members, $closeBrace] = wasm.ast.Selectors.sequence(this.cst);
            return {
                abstractKeyword: $abstractKeyword === undefined ? undefined : $abstractKeyword,
                contractKeyword: $contractKeyword,
                name: $name,
                specifiers: new ContractSpecifiers($specifiers),
                openBrace: $openBrace,
                members: new ContractMembers($members),
                closeBrace: $closeBrace,
            };
        });
        assertKind(this.cst.kind, NonterminalKind.ContractDefinition);
    }
    /**
     * Returns the child node that has the label `abstract_keyword`.
     */
    get abstractKeyword() {
        return this.fetch().abstractKeyword;
    }
    /**
     * Returns the child node that has the label `contract_keyword`.
     */
    get contractKeyword() {
        return this.fetch().contractKeyword;
    }
    /**
     * Returns the child node that has the label `name`.
     */
    get name() {
        return this.fetch().name;
    }
    /**
     * Returns the child node that has the label `specifiers`.
     */
    get specifiers() {
        return this.fetch().specifiers;
    }
    /**
     * Returns the child node that has the label `open_brace`.
     */
    get openBrace() {
        return this.fetch().openBrace;
    }
    /**
     * Returns the child node that has the label `members`.
     */
    get members() {
        return this.fetch().members;
    }
    /**
     * Returns the child node that has the label `close_brace`.
     */
    get closeBrace() {
        return this.fetch().closeBrace;
    }
}
/**
 * This node represents a `InheritanceSpecifier` nonterminal, with the following structure:
 *
 * ```ebnf
 * InheritanceSpecifier = (* is_keyword: *) IS_KEYWORD
 *                        (* types: *) InheritanceTypes;
 * ```
 */
export class InheritanceSpecifier {
    /**
     * Constructs a new AST node of type `InheritanceSpecifier`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `InheritanceSpecifier`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const [$isKeyword, $types] = wasm.ast.Selectors.sequence(this.cst);
            return {
                isKeyword: $isKeyword,
                types: new InheritanceTypes($types),
            };
        });
        assertKind(this.cst.kind, NonterminalKind.InheritanceSpecifier);
    }
    /**
     * Returns the child node that has the label `is_keyword`.
     */
    get isKeyword() {
        return this.fetch().isKeyword;
    }
    /**
     * Returns the child node that has the label `types`.
     */
    get types() {
        return this.fetch().types;
    }
}
/**
 * This node represents a `InheritanceType` nonterminal, with the following structure:
 *
 * ```ebnf
 * InheritanceType = (* type_name: *) IdentifierPath
 *                   (* arguments: *) ArgumentsDeclaration?;
 * ```
 */
export class InheritanceType {
    /**
     * Constructs a new AST node of type `InheritanceType`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `InheritanceType`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const [$typeName, $arguments] = wasm.ast.Selectors.sequence(this.cst);
            return {
                typeName: new IdentifierPath($typeName),
                arguments: $arguments === undefined ? undefined : new ArgumentsDeclaration($arguments),
            };
        });
        assertKind(this.cst.kind, NonterminalKind.InheritanceType);
    }
    /**
     * Returns the child node that has the label `type_name`.
     */
    get typeName() {
        return this.fetch().typeName;
    }
    /**
     * Returns the child node that has the label `arguments`.
     */
    get arguments() {
        return this.fetch().arguments;
    }
}
/**
 * This node represents a `StorageLayoutSpecifier` nonterminal, with the following structure:
 *
 * ```ebnf
 * (* Introduced in 0.8.29 *)
 * StorageLayoutSpecifier = (* layout_keyword: *) LAYOUT_KEYWORD
 *                          (* at_keyword: *) AT_KEYWORD
 *                          (* expression: *) Expression;
 * ```
 */
export class StorageLayoutSpecifier {
    /**
     * Constructs a new AST node of type `StorageLayoutSpecifier`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `StorageLayoutSpecifier`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const [$layoutKeyword, $atKeyword, $expression] = wasm.ast.Selectors.sequence(this.cst);
            return {
                layoutKeyword: $layoutKeyword,
                atKeyword: $atKeyword,
                expression: new Expression($expression),
            };
        });
        assertKind(this.cst.kind, NonterminalKind.StorageLayoutSpecifier);
    }
    /**
     * Returns the child node that has the label `layout_keyword`.
     */
    get layoutKeyword() {
        return this.fetch().layoutKeyword;
    }
    /**
     * Returns the child node that has the label `at_keyword`.
     */
    get atKeyword() {
        return this.fetch().atKeyword;
    }
    /**
     * Returns the child node that has the label `expression`.
     */
    get expression() {
        return this.fetch().expression;
    }
}
/**
 * This node represents a `InterfaceDefinition` nonterminal, with the following structure:
 *
 * ```ebnf
 * InterfaceDefinition = (* interface_keyword: *) INTERFACE_KEYWORD
 *                       (* name: *) IDENTIFIER
 *                       (* inheritance: *) InheritanceSpecifier?
 *                       (* open_brace: *) OPEN_BRACE
 *                       (* members: *) InterfaceMembers
 *                       (* close_brace: *) CLOSE_BRACE;
 * ```
 */
export class InterfaceDefinition {
    /**
     * Constructs a new AST node of type `InterfaceDefinition`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `InterfaceDefinition`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const [$interfaceKeyword, $name, $inheritance, $openBrace, $members, $closeBrace] = wasm.ast.Selectors.sequence(this.cst);
            return {
                interfaceKeyword: $interfaceKeyword,
                name: $name,
                inheritance: $inheritance === undefined ? undefined : new InheritanceSpecifier($inheritance),
                openBrace: $openBrace,
                members: new InterfaceMembers($members),
                closeBrace: $closeBrace,
            };
        });
        assertKind(this.cst.kind, NonterminalKind.InterfaceDefinition);
    }
    /**
     * Returns the child node that has the label `interface_keyword`.
     */
    get interfaceKeyword() {
        return this.fetch().interfaceKeyword;
    }
    /**
     * Returns the child node that has the label `name`.
     */
    get name() {
        return this.fetch().name;
    }
    /**
     * Returns the child node that has the label `inheritance`.
     */
    get inheritance() {
        return this.fetch().inheritance;
    }
    /**
     * Returns the child node that has the label `open_brace`.
     */
    get openBrace() {
        return this.fetch().openBrace;
    }
    /**
     * Returns the child node that has the label `members`.
     */
    get members() {
        return this.fetch().members;
    }
    /**
     * Returns the child node that has the label `close_brace`.
     */
    get closeBrace() {
        return this.fetch().closeBrace;
    }
}
/**
 * This node represents a `LibraryDefinition` nonterminal, with the following structure:
 *
 * ```ebnf
 * LibraryDefinition = (* library_keyword: *) LIBRARY_KEYWORD
 *                     (* name: *) IDENTIFIER
 *                     (* open_brace: *) OPEN_BRACE
 *                     (* members: *) LibraryMembers
 *                     (* close_brace: *) CLOSE_BRACE;
 * ```
 */
export class LibraryDefinition {
    /**
     * Constructs a new AST node of type `LibraryDefinition`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `LibraryDefinition`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const [$libraryKeyword, $name, $openBrace, $members, $closeBrace] = wasm.ast.Selectors.sequence(this.cst);
            return {
                libraryKeyword: $libraryKeyword,
                name: $name,
                openBrace: $openBrace,
                members: new LibraryMembers($members),
                closeBrace: $closeBrace,
            };
        });
        assertKind(this.cst.kind, NonterminalKind.LibraryDefinition);
    }
    /**
     * Returns the child node that has the label `library_keyword`.
     */
    get libraryKeyword() {
        return this.fetch().libraryKeyword;
    }
    /**
     * Returns the child node that has the label `name`.
     */
    get name() {
        return this.fetch().name;
    }
    /**
     * Returns the child node that has the label `open_brace`.
     */
    get openBrace() {
        return this.fetch().openBrace;
    }
    /**
     * Returns the child node that has the label `members`.
     */
    get members() {
        return this.fetch().members;
    }
    /**
     * Returns the child node that has the label `close_brace`.
     */
    get closeBrace() {
        return this.fetch().closeBrace;
    }
}
/**
 * This node represents a `StructDefinition` nonterminal, with the following structure:
 *
 * ```ebnf
 * StructDefinition = (* struct_keyword: *) STRUCT_KEYWORD
 *                    (* name: *) IDENTIFIER
 *                    (* open_brace: *) OPEN_BRACE
 *                    (* members: *) StructMembers
 *                    (* close_brace: *) CLOSE_BRACE;
 * ```
 */
export class StructDefinition {
    /**
     * Constructs a new AST node of type `StructDefinition`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `StructDefinition`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const [$structKeyword, $name, $openBrace, $members, $closeBrace] = wasm.ast.Selectors.sequence(this.cst);
            return {
                structKeyword: $structKeyword,
                name: $name,
                openBrace: $openBrace,
                members: new StructMembers($members),
                closeBrace: $closeBrace,
            };
        });
        assertKind(this.cst.kind, NonterminalKind.StructDefinition);
    }
    /**
     * Returns the child node that has the label `struct_keyword`.
     */
    get structKeyword() {
        return this.fetch().structKeyword;
    }
    /**
     * Returns the child node that has the label `name`.
     */
    get name() {
        return this.fetch().name;
    }
    /**
     * Returns the child node that has the label `open_brace`.
     */
    get openBrace() {
        return this.fetch().openBrace;
    }
    /**
     * Returns the child node that has the label `members`.
     */
    get members() {
        return this.fetch().members;
    }
    /**
     * Returns the child node that has the label `close_brace`.
     */
    get closeBrace() {
        return this.fetch().closeBrace;
    }
}
/**
 * This node represents a `StructMember` nonterminal, with the following structure:
 *
 * ```ebnf
 * StructMember = (* type_name: *) TypeName
 *                (* name: *) IDENTIFIER
 *                (* semicolon: *) SEMICOLON;
 * ```
 */
export class StructMember {
    /**
     * Constructs a new AST node of type `StructMember`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `StructMember`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const [$typeName, $name, $semicolon] = wasm.ast.Selectors.sequence(this.cst);
            return {
                typeName: new TypeName($typeName),
                name: $name,
                semicolon: $semicolon,
            };
        });
        assertKind(this.cst.kind, NonterminalKind.StructMember);
    }
    /**
     * Returns the child node that has the label `type_name`.
     */
    get typeName() {
        return this.fetch().typeName;
    }
    /**
     * Returns the child node that has the label `name`.
     */
    get name() {
        return this.fetch().name;
    }
    /**
     * Returns the child node that has the label `semicolon`.
     */
    get semicolon() {
        return this.fetch().semicolon;
    }
}
/**
 * This node represents a `EnumDefinition` nonterminal, with the following structure:
 *
 * ```ebnf
 * EnumDefinition = (* enum_keyword: *) ENUM_KEYWORD
 *                  (* name: *) IDENTIFIER
 *                  (* open_brace: *) OPEN_BRACE
 *                  (* members: *) EnumMembers
 *                  (* close_brace: *) CLOSE_BRACE;
 * ```
 */
export class EnumDefinition {
    /**
     * Constructs a new AST node of type `EnumDefinition`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `EnumDefinition`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const [$enumKeyword, $name, $openBrace, $members, $closeBrace] = wasm.ast.Selectors.sequence(this.cst);
            return {
                enumKeyword: $enumKeyword,
                name: $name,
                openBrace: $openBrace,
                members: new EnumMembers($members),
                closeBrace: $closeBrace,
            };
        });
        assertKind(this.cst.kind, NonterminalKind.EnumDefinition);
    }
    /**
     * Returns the child node that has the label `enum_keyword`.
     */
    get enumKeyword() {
        return this.fetch().enumKeyword;
    }
    /**
     * Returns the child node that has the label `name`.
     */
    get name() {
        return this.fetch().name;
    }
    /**
     * Returns the child node that has the label `open_brace`.
     */
    get openBrace() {
        return this.fetch().openBrace;
    }
    /**
     * Returns the child node that has the label `members`.
     */
    get members() {
        return this.fetch().members;
    }
    /**
     * Returns the child node that has the label `close_brace`.
     */
    get closeBrace() {
        return this.fetch().closeBrace;
    }
}
/**
 * This node represents a `ConstantDefinition` nonterminal, with the following structure:
 *
 * ```ebnf
 * (* Introduced in 0.7.4 *)
 * ConstantDefinition = (* type_name: *) TypeName
 *                      (* constant_keyword: *) CONSTANT_KEYWORD
 *                      (* name: *) IDENTIFIER
 *                      (* equal: *) EQUAL
 *                      (* value: *) Expression
 *                      (* semicolon: *) SEMICOLON;
 * ```
 */
export class ConstantDefinition {
    /**
     * Constructs a new AST node of type `ConstantDefinition`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `ConstantDefinition`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const [$typeName, $constantKeyword, $name, $equal, $value, $semicolon] = wasm.ast.Selectors.sequence(this.cst);
            return {
                typeName: new TypeName($typeName),
                constantKeyword: $constantKeyword,
                name: $name,
                equal: $equal,
                value: new Expression($value),
                semicolon: $semicolon,
            };
        });
        assertKind(this.cst.kind, NonterminalKind.ConstantDefinition);
    }
    /**
     * Returns the child node that has the label `type_name`.
     */
    get typeName() {
        return this.fetch().typeName;
    }
    /**
     * Returns the child node that has the label `constant_keyword`.
     */
    get constantKeyword() {
        return this.fetch().constantKeyword;
    }
    /**
     * Returns the child node that has the label `name`.
     */
    get name() {
        return this.fetch().name;
    }
    /**
     * Returns the child node that has the label `equal`.
     */
    get equal() {
        return this.fetch().equal;
    }
    /**
     * Returns the child node that has the label `value`.
     */
    get value() {
        return this.fetch().value;
    }
    /**
     * Returns the child node that has the label `semicolon`.
     */
    get semicolon() {
        return this.fetch().semicolon;
    }
}
/**
 * This node represents a `StateVariableDefinition` nonterminal, with the following structure:
 *
 * ```ebnf
 * StateVariableDefinition = (* type_name: *) TypeName
 *                           (* attributes: *) StateVariableAttributes
 *                           (* name: *) IDENTIFIER
 *                           (* value: *) StateVariableDefinitionValue?
 *                           (* semicolon: *) SEMICOLON;
 * ```
 */
export class StateVariableDefinition {
    /**
     * Constructs a new AST node of type `StateVariableDefinition`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `StateVariableDefinition`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const [$typeName, $attributes, $name, $value, $semicolon] = wasm.ast.Selectors.sequence(this.cst);
            return {
                typeName: new TypeName($typeName),
                attributes: new StateVariableAttributes($attributes),
                name: $name,
                value: $value === undefined ? undefined : new StateVariableDefinitionValue($value),
                semicolon: $semicolon,
            };
        });
        assertKind(this.cst.kind, NonterminalKind.StateVariableDefinition);
    }
    /**
     * Returns the child node that has the label `type_name`.
     */
    get typeName() {
        return this.fetch().typeName;
    }
    /**
     * Returns the child node that has the label `attributes`.
     */
    get attributes() {
        return this.fetch().attributes;
    }
    /**
     * Returns the child node that has the label `name`.
     */
    get name() {
        return this.fetch().name;
    }
    /**
     * Returns the child node that has the label `value`.
     */
    get value() {
        return this.fetch().value;
    }
    /**
     * Returns the child node that has the label `semicolon`.
     */
    get semicolon() {
        return this.fetch().semicolon;
    }
}
/**
 * This node represents a `StateVariableDefinitionValue` nonterminal, with the following structure:
 *
 * ```ebnf
 * StateVariableDefinitionValue = (* equal: *) EQUAL
 *                                (* value: *) Expression;
 * ```
 */
export class StateVariableDefinitionValue {
    /**
     * Constructs a new AST node of type `StateVariableDefinitionValue`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `StateVariableDefinitionValue`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const [$equal, $value] = wasm.ast.Selectors.sequence(this.cst);
            return {
                equal: $equal,
                value: new Expression($value),
            };
        });
        assertKind(this.cst.kind, NonterminalKind.StateVariableDefinitionValue);
    }
    /**
     * Returns the child node that has the label `equal`.
     */
    get equal() {
        return this.fetch().equal;
    }
    /**
     * Returns the child node that has the label `value`.
     */
    get value() {
        return this.fetch().value;
    }
}
/**
 * This node represents a `FunctionDefinition` nonterminal, with the following structure:
 *
 * ```ebnf
 * FunctionDefinition = (* function_keyword: *) FUNCTION_KEYWORD
 *                      (* name: *) FunctionName
 *                      (* parameters: *) ParametersDeclaration
 *                      (* attributes: *) FunctionAttributes
 *                      (* returns: *) ReturnsDeclaration?
 *                      (* body: *) FunctionBody;
 * ```
 */
export class FunctionDefinition {
    /**
     * Constructs a new AST node of type `FunctionDefinition`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `FunctionDefinition`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const [$functionKeyword, $name, $parameters, $attributes, $returns, $body] = wasm.ast.Selectors.sequence(this.cst);
            return {
                functionKeyword: $functionKeyword,
                name: new FunctionName($name),
                parameters: new ParametersDeclaration($parameters),
                attributes: new FunctionAttributes($attributes),
                returns: $returns === undefined ? undefined : new ReturnsDeclaration($returns),
                body: new FunctionBody($body),
            };
        });
        assertKind(this.cst.kind, NonterminalKind.FunctionDefinition);
    }
    /**
     * Returns the child node that has the label `function_keyword`.
     */
    get functionKeyword() {
        return this.fetch().functionKeyword;
    }
    /**
     * Returns the child node that has the label `name`.
     */
    get name() {
        return this.fetch().name;
    }
    /**
     * Returns the child node that has the label `parameters`.
     */
    get parameters() {
        return this.fetch().parameters;
    }
    /**
     * Returns the child node that has the label `attributes`.
     */
    get attributes() {
        return this.fetch().attributes;
    }
    /**
     * Returns the child node that has the label `returns`.
     */
    get returns() {
        return this.fetch().returns;
    }
    /**
     * Returns the child node that has the label `body`.
     */
    get body() {
        return this.fetch().body;
    }
}
/**
 * This node represents a `ParametersDeclaration` nonterminal, with the following structure:
 *
 * ```ebnf
 * ParametersDeclaration = (* open_paren: *) OPEN_PAREN
 *                         (* parameters: *) Parameters
 *                         (* close_paren: *) CLOSE_PAREN;
 * ```
 */
export class ParametersDeclaration {
    /**
     * Constructs a new AST node of type `ParametersDeclaration`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `ParametersDeclaration`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const [$openParen, $parameters, $closeParen] = wasm.ast.Selectors.sequence(this.cst);
            return {
                openParen: $openParen,
                parameters: new Parameters($parameters),
                closeParen: $closeParen,
            };
        });
        assertKind(this.cst.kind, NonterminalKind.ParametersDeclaration);
    }
    /**
     * Returns the child node that has the label `open_paren`.
     */
    get openParen() {
        return this.fetch().openParen;
    }
    /**
     * Returns the child node that has the label `parameters`.
     */
    get parameters() {
        return this.fetch().parameters;
    }
    /**
     * Returns the child node that has the label `close_paren`.
     */
    get closeParen() {
        return this.fetch().closeParen;
    }
}
/**
 * This node represents a `Parameter` nonterminal, with the following structure:
 *
 * ```ebnf
 * Parameter = (* type_name: *) TypeName
 *             (* storage_location: *) StorageLocation?
 *             (* name: *) IDENTIFIER?;
 * ```
 */
export class Parameter {
    /**
     * Constructs a new AST node of type `Parameter`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `Parameter`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const [$typeName, $storageLocation, $name] = wasm.ast.Selectors.sequence(this.cst);
            return {
                typeName: new TypeName($typeName),
                storageLocation: $storageLocation === undefined ? undefined : new StorageLocation($storageLocation),
                name: $name === undefined ? undefined : $name,
            };
        });
        assertKind(this.cst.kind, NonterminalKind.Parameter);
    }
    /**
     * Returns the child node that has the label `type_name`.
     */
    get typeName() {
        return this.fetch().typeName;
    }
    /**
     * Returns the child node that has the label `storage_location`.
     */
    get storageLocation() {
        return this.fetch().storageLocation;
    }
    /**
     * Returns the child node that has the label `name`.
     */
    get name() {
        return this.fetch().name;
    }
}
/**
 * This node represents a `OverrideSpecifier` nonterminal, with the following structure:
 *
 * ```ebnf
 * (* Introduced in 0.6.0 *)
 * OverrideSpecifier = (* override_keyword: *) OVERRIDE_KEYWORD
 *                     (* overridden: *) OverridePathsDeclaration?;
 * ```
 */
export class OverrideSpecifier {
    /**
     * Constructs a new AST node of type `OverrideSpecifier`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `OverrideSpecifier`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const [$overrideKeyword, $overridden] = wasm.ast.Selectors.sequence(this.cst);
            return {
                overrideKeyword: $overrideKeyword,
                overridden: $overridden === undefined ? undefined : new OverridePathsDeclaration($overridden),
            };
        });
        assertKind(this.cst.kind, NonterminalKind.OverrideSpecifier);
    }
    /**
     * Returns the child node that has the label `override_keyword`.
     */
    get overrideKeyword() {
        return this.fetch().overrideKeyword;
    }
    /**
     * Returns the child node that has the label `overridden`.
     */
    get overridden() {
        return this.fetch().overridden;
    }
}
/**
 * This node represents a `OverridePathsDeclaration` nonterminal, with the following structure:
 *
 * ```ebnf
 * (* Introduced in 0.6.0 *)
 * OverridePathsDeclaration = (* open_paren: *) OPEN_PAREN
 *                            (* paths: *) OverridePaths
 *                            (* close_paren: *) CLOSE_PAREN;
 * ```
 */
export class OverridePathsDeclaration {
    /**
     * Constructs a new AST node of type `OverridePathsDeclaration`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `OverridePathsDeclaration`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const [$openParen, $paths, $closeParen] = wasm.ast.Selectors.sequence(this.cst);
            return {
                openParen: $openParen,
                paths: new OverridePaths($paths),
                closeParen: $closeParen,
            };
        });
        assertKind(this.cst.kind, NonterminalKind.OverridePathsDeclaration);
    }
    /**
     * Returns the child node that has the label `open_paren`.
     */
    get openParen() {
        return this.fetch().openParen;
    }
    /**
     * Returns the child node that has the label `paths`.
     */
    get paths() {
        return this.fetch().paths;
    }
    /**
     * Returns the child node that has the label `close_paren`.
     */
    get closeParen() {
        return this.fetch().closeParen;
    }
}
/**
 * This node represents a `ReturnsDeclaration` nonterminal, with the following structure:
 *
 * ```ebnf
 * ReturnsDeclaration = (* returns_keyword: *) RETURNS_KEYWORD
 *                      (* variables: *) ParametersDeclaration;
 * ```
 */
export class ReturnsDeclaration {
    /**
     * Constructs a new AST node of type `ReturnsDeclaration`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `ReturnsDeclaration`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const [$returnsKeyword, $variables] = wasm.ast.Selectors.sequence(this.cst);
            return {
                returnsKeyword: $returnsKeyword,
                variables: new ParametersDeclaration($variables),
            };
        });
        assertKind(this.cst.kind, NonterminalKind.ReturnsDeclaration);
    }
    /**
     * Returns the child node that has the label `returns_keyword`.
     */
    get returnsKeyword() {
        return this.fetch().returnsKeyword;
    }
    /**
     * Returns the child node that has the label `variables`.
     */
    get variables() {
        return this.fetch().variables;
    }
}
/**
 * This node represents a `ConstructorDefinition` nonterminal, with the following structure:
 *
 * ```ebnf
 * (* Introduced in 0.4.22 *)
 * ConstructorDefinition = (* constructor_keyword: *) CONSTRUCTOR_KEYWORD
 *                         (* parameters: *) ParametersDeclaration
 *                         (* attributes: *) ConstructorAttributes
 *                         (* body: *) Block;
 * ```
 */
export class ConstructorDefinition {
    /**
     * Constructs a new AST node of type `ConstructorDefinition`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `ConstructorDefinition`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const [$constructorKeyword, $parameters, $attributes, $body] = wasm.ast.Selectors.sequence(this.cst);
            return {
                constructorKeyword: $constructorKeyword,
                parameters: new ParametersDeclaration($parameters),
                attributes: new ConstructorAttributes($attributes),
                body: new Block($body),
            };
        });
        assertKind(this.cst.kind, NonterminalKind.ConstructorDefinition);
    }
    /**
     * Returns the child node that has the label `constructor_keyword`.
     */
    get constructorKeyword() {
        return this.fetch().constructorKeyword;
    }
    /**
     * Returns the child node that has the label `parameters`.
     */
    get parameters() {
        return this.fetch().parameters;
    }
    /**
     * Returns the child node that has the label `attributes`.
     */
    get attributes() {
        return this.fetch().attributes;
    }
    /**
     * Returns the child node that has the label `body`.
     */
    get body() {
        return this.fetch().body;
    }
}
/**
 * This node represents a `UnnamedFunctionDefinition` nonterminal, with the following structure:
 *
 * ```ebnf
 * (* Deprecated in 0.6.0 *)
 * UnnamedFunctionDefinition = (* function_keyword: *) FUNCTION_KEYWORD
 *                             (* parameters: *) ParametersDeclaration
 *                             (* attributes: *) UnnamedFunctionAttributes
 *                             (* body: *) FunctionBody;
 * ```
 */
export class UnnamedFunctionDefinition {
    /**
     * Constructs a new AST node of type `UnnamedFunctionDefinition`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `UnnamedFunctionDefinition`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const [$functionKeyword, $parameters, $attributes, $body] = wasm.ast.Selectors.sequence(this.cst);
            return {
                functionKeyword: $functionKeyword,
                parameters: new ParametersDeclaration($parameters),
                attributes: new UnnamedFunctionAttributes($attributes),
                body: new FunctionBody($body),
            };
        });
        assertKind(this.cst.kind, NonterminalKind.UnnamedFunctionDefinition);
    }
    /**
     * Returns the child node that has the label `function_keyword`.
     */
    get functionKeyword() {
        return this.fetch().functionKeyword;
    }
    /**
     * Returns the child node that has the label `parameters`.
     */
    get parameters() {
        return this.fetch().parameters;
    }
    /**
     * Returns the child node that has the label `attributes`.
     */
    get attributes() {
        return this.fetch().attributes;
    }
    /**
     * Returns the child node that has the label `body`.
     */
    get body() {
        return this.fetch().body;
    }
}
/**
 * This node represents a `FallbackFunctionDefinition` nonterminal, with the following structure:
 *
 * ```ebnf
 * (* Introduced in 0.6.0 *)
 * FallbackFunctionDefinition = (* fallback_keyword: *) FALLBACK_KEYWORD
 *                              (* parameters: *) ParametersDeclaration
 *                              (* attributes: *) FallbackFunctionAttributes
 *                              (* returns: *) ReturnsDeclaration?
 *                              (* body: *) FunctionBody;
 * ```
 */
export class FallbackFunctionDefinition {
    /**
     * Constructs a new AST node of type `FallbackFunctionDefinition`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `FallbackFunctionDefinition`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const [$fallbackKeyword, $parameters, $attributes, $returns, $body] = wasm.ast.Selectors.sequence(this.cst);
            return {
                fallbackKeyword: $fallbackKeyword,
                parameters: new ParametersDeclaration($parameters),
                attributes: new FallbackFunctionAttributes($attributes),
                returns: $returns === undefined ? undefined : new ReturnsDeclaration($returns),
                body: new FunctionBody($body),
            };
        });
        assertKind(this.cst.kind, NonterminalKind.FallbackFunctionDefinition);
    }
    /**
     * Returns the child node that has the label `fallback_keyword`.
     */
    get fallbackKeyword() {
        return this.fetch().fallbackKeyword;
    }
    /**
     * Returns the child node that has the label `parameters`.
     */
    get parameters() {
        return this.fetch().parameters;
    }
    /**
     * Returns the child node that has the label `attributes`.
     */
    get attributes() {
        return this.fetch().attributes;
    }
    /**
     * Returns the child node that has the label `returns`.
     */
    get returns() {
        return this.fetch().returns;
    }
    /**
     * Returns the child node that has the label `body`.
     */
    get body() {
        return this.fetch().body;
    }
}
/**
 * This node represents a `ReceiveFunctionDefinition` nonterminal, with the following structure:
 *
 * ```ebnf
 * (* Introduced in 0.6.0 *)
 * ReceiveFunctionDefinition = (* receive_keyword: *) RECEIVE_KEYWORD
 *                             (* parameters: *) ParametersDeclaration
 *                             (* attributes: *) ReceiveFunctionAttributes
 *                             (* body: *) FunctionBody;
 * ```
 */
export class ReceiveFunctionDefinition {
    /**
     * Constructs a new AST node of type `ReceiveFunctionDefinition`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `ReceiveFunctionDefinition`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const [$receiveKeyword, $parameters, $attributes, $body] = wasm.ast.Selectors.sequence(this.cst);
            return {
                receiveKeyword: $receiveKeyword,
                parameters: new ParametersDeclaration($parameters),
                attributes: new ReceiveFunctionAttributes($attributes),
                body: new FunctionBody($body),
            };
        });
        assertKind(this.cst.kind, NonterminalKind.ReceiveFunctionDefinition);
    }
    /**
     * Returns the child node that has the label `receive_keyword`.
     */
    get receiveKeyword() {
        return this.fetch().receiveKeyword;
    }
    /**
     * Returns the child node that has the label `parameters`.
     */
    get parameters() {
        return this.fetch().parameters;
    }
    /**
     * Returns the child node that has the label `attributes`.
     */
    get attributes() {
        return this.fetch().attributes;
    }
    /**
     * Returns the child node that has the label `body`.
     */
    get body() {
        return this.fetch().body;
    }
}
/**
 * This node represents a `ModifierDefinition` nonterminal, with the following structure:
 *
 * ```ebnf
 * ModifierDefinition = (* modifier_keyword: *) MODIFIER_KEYWORD
 *                      (* name: *) IDENTIFIER
 *                      (* parameters: *) ParametersDeclaration?
 *                      (* attributes: *) ModifierAttributes
 *                      (* body: *) FunctionBody;
 * ```
 */
export class ModifierDefinition {
    /**
     * Constructs a new AST node of type `ModifierDefinition`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `ModifierDefinition`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const [$modifierKeyword, $name, $parameters, $attributes, $body] = wasm.ast.Selectors.sequence(this.cst);
            return {
                modifierKeyword: $modifierKeyword,
                name: $name,
                parameters: $parameters === undefined ? undefined : new ParametersDeclaration($parameters),
                attributes: new ModifierAttributes($attributes),
                body: new FunctionBody($body),
            };
        });
        assertKind(this.cst.kind, NonterminalKind.ModifierDefinition);
    }
    /**
     * Returns the child node that has the label `modifier_keyword`.
     */
    get modifierKeyword() {
        return this.fetch().modifierKeyword;
    }
    /**
     * Returns the child node that has the label `name`.
     */
    get name() {
        return this.fetch().name;
    }
    /**
     * Returns the child node that has the label `parameters`.
     */
    get parameters() {
        return this.fetch().parameters;
    }
    /**
     * Returns the child node that has the label `attributes`.
     */
    get attributes() {
        return this.fetch().attributes;
    }
    /**
     * Returns the child node that has the label `body`.
     */
    get body() {
        return this.fetch().body;
    }
}
/**
 * This node represents a `ModifierInvocation` nonterminal, with the following structure:
 *
 * ```ebnf
 * ModifierInvocation = (* name: *) IdentifierPath
 *                      (* arguments: *) ArgumentsDeclaration?;
 * ```
 */
export class ModifierInvocation {
    /**
     * Constructs a new AST node of type `ModifierInvocation`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `ModifierInvocation`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const [$name, $arguments] = wasm.ast.Selectors.sequence(this.cst);
            return {
                name: new IdentifierPath($name),
                arguments: $arguments === undefined ? undefined : new ArgumentsDeclaration($arguments),
            };
        });
        assertKind(this.cst.kind, NonterminalKind.ModifierInvocation);
    }
    /**
     * Returns the child node that has the label `name`.
     */
    get name() {
        return this.fetch().name;
    }
    /**
     * Returns the child node that has the label `arguments`.
     */
    get arguments() {
        return this.fetch().arguments;
    }
}
/**
 * This node represents a `EventDefinition` nonterminal, with the following structure:
 *
 * ```ebnf
 * EventDefinition = (* event_keyword: *) EVENT_KEYWORD
 *                   (* name: *) IDENTIFIER
 *                   (* parameters: *) EventParametersDeclaration
 *                   (* anonymous_keyword: *) ANONYMOUS_KEYWORD?
 *                   (* semicolon: *) SEMICOLON;
 * ```
 */
export class EventDefinition {
    /**
     * Constructs a new AST node of type `EventDefinition`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `EventDefinition`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const [$eventKeyword, $name, $parameters, $anonymousKeyword, $semicolon] = wasm.ast.Selectors.sequence(this.cst);
            return {
                eventKeyword: $eventKeyword,
                name: $name,
                parameters: new EventParametersDeclaration($parameters),
                anonymousKeyword: $anonymousKeyword === undefined ? undefined : $anonymousKeyword,
                semicolon: $semicolon,
            };
        });
        assertKind(this.cst.kind, NonterminalKind.EventDefinition);
    }
    /**
     * Returns the child node that has the label `event_keyword`.
     */
    get eventKeyword() {
        return this.fetch().eventKeyword;
    }
    /**
     * Returns the child node that has the label `name`.
     */
    get name() {
        return this.fetch().name;
    }
    /**
     * Returns the child node that has the label `parameters`.
     */
    get parameters() {
        return this.fetch().parameters;
    }
    /**
     * Returns the child node that has the label `anonymous_keyword`.
     */
    get anonymousKeyword() {
        return this.fetch().anonymousKeyword;
    }
    /**
     * Returns the child node that has the label `semicolon`.
     */
    get semicolon() {
        return this.fetch().semicolon;
    }
}
/**
 * This node represents a `EventParametersDeclaration` nonterminal, with the following structure:
 *
 * ```ebnf
 * EventParametersDeclaration = (* open_paren: *) OPEN_PAREN
 *                              (* parameters: *) EventParameters
 *                              (* close_paren: *) CLOSE_PAREN;
 * ```
 */
export class EventParametersDeclaration {
    /**
     * Constructs a new AST node of type `EventParametersDeclaration`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `EventParametersDeclaration`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const [$openParen, $parameters, $closeParen] = wasm.ast.Selectors.sequence(this.cst);
            return {
                openParen: $openParen,
                parameters: new EventParameters($parameters),
                closeParen: $closeParen,
            };
        });
        assertKind(this.cst.kind, NonterminalKind.EventParametersDeclaration);
    }
    /**
     * Returns the child node that has the label `open_paren`.
     */
    get openParen() {
        return this.fetch().openParen;
    }
    /**
     * Returns the child node that has the label `parameters`.
     */
    get parameters() {
        return this.fetch().parameters;
    }
    /**
     * Returns the child node that has the label `close_paren`.
     */
    get closeParen() {
        return this.fetch().closeParen;
    }
}
/**
 * This node represents a `EventParameter` nonterminal, with the following structure:
 *
 * ```ebnf
 * EventParameter = (* type_name: *) TypeName
 *                  (* indexed_keyword: *) INDEXED_KEYWORD?
 *                  (* name: *) IDENTIFIER?;
 * ```
 */
export class EventParameter {
    /**
     * Constructs a new AST node of type `EventParameter`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `EventParameter`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const [$typeName, $indexedKeyword, $name] = wasm.ast.Selectors.sequence(this.cst);
            return {
                typeName: new TypeName($typeName),
                indexedKeyword: $indexedKeyword === undefined ? undefined : $indexedKeyword,
                name: $name === undefined ? undefined : $name,
            };
        });
        assertKind(this.cst.kind, NonterminalKind.EventParameter);
    }
    /**
     * Returns the child node that has the label `type_name`.
     */
    get typeName() {
        return this.fetch().typeName;
    }
    /**
     * Returns the child node that has the label `indexed_keyword`.
     */
    get indexedKeyword() {
        return this.fetch().indexedKeyword;
    }
    /**
     * Returns the child node that has the label `name`.
     */
    get name() {
        return this.fetch().name;
    }
}
/**
 * This node represents a `UserDefinedValueTypeDefinition` nonterminal, with the following structure:
 *
 * ```ebnf
 * (* Introduced in 0.8.8 *)
 * UserDefinedValueTypeDefinition = (* type_keyword: *) TYPE_KEYWORD
 *                                  (* name: *) IDENTIFIER
 *                                  (* is_keyword: *) IS_KEYWORD
 *                                  (* value_type: *) ElementaryType
 *                                  (* semicolon: *) SEMICOLON;
 * ```
 */
export class UserDefinedValueTypeDefinition {
    /**
     * Constructs a new AST node of type `UserDefinedValueTypeDefinition`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `UserDefinedValueTypeDefinition`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const [$typeKeyword, $name, $isKeyword, $valueType, $semicolon] = wasm.ast.Selectors.sequence(this.cst);
            return {
                typeKeyword: $typeKeyword,
                name: $name,
                isKeyword: $isKeyword,
                valueType: new ElementaryType($valueType),
                semicolon: $semicolon,
            };
        });
        assertKind(this.cst.kind, NonterminalKind.UserDefinedValueTypeDefinition);
    }
    /**
     * Returns the child node that has the label `type_keyword`.
     */
    get typeKeyword() {
        return this.fetch().typeKeyword;
    }
    /**
     * Returns the child node that has the label `name`.
     */
    get name() {
        return this.fetch().name;
    }
    /**
     * Returns the child node that has the label `is_keyword`.
     */
    get isKeyword() {
        return this.fetch().isKeyword;
    }
    /**
     * Returns the child node that has the label `value_type`.
     */
    get valueType() {
        return this.fetch().valueType;
    }
    /**
     * Returns the child node that has the label `semicolon`.
     */
    get semicolon() {
        return this.fetch().semicolon;
    }
}
/**
 * This node represents a `ErrorDefinition` nonterminal, with the following structure:
 *
 * ```ebnf
 * (* Introduced in 0.8.4 *)
 * ErrorDefinition = (* error_keyword: *) ERROR_KEYWORD
 *                   (* name: *) IDENTIFIER
 *                   (* members: *) ErrorParametersDeclaration
 *                   (* semicolon: *) SEMICOLON;
 * ```
 */
export class ErrorDefinition {
    /**
     * Constructs a new AST node of type `ErrorDefinition`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `ErrorDefinition`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const [$errorKeyword, $name, $members, $semicolon] = wasm.ast.Selectors.sequence(this.cst);
            return {
                errorKeyword: $errorKeyword,
                name: $name,
                members: new ErrorParametersDeclaration($members),
                semicolon: $semicolon,
            };
        });
        assertKind(this.cst.kind, NonterminalKind.ErrorDefinition);
    }
    /**
     * Returns the child node that has the label `error_keyword`.
     */
    get errorKeyword() {
        return this.fetch().errorKeyword;
    }
    /**
     * Returns the child node that has the label `name`.
     */
    get name() {
        return this.fetch().name;
    }
    /**
     * Returns the child node that has the label `members`.
     */
    get members() {
        return this.fetch().members;
    }
    /**
     * Returns the child node that has the label `semicolon`.
     */
    get semicolon() {
        return this.fetch().semicolon;
    }
}
/**
 * This node represents a `ErrorParametersDeclaration` nonterminal, with the following structure:
 *
 * ```ebnf
 * (* Introduced in 0.8.4 *)
 * ErrorParametersDeclaration = (* open_paren: *) OPEN_PAREN
 *                              (* parameters: *) ErrorParameters
 *                              (* close_paren: *) CLOSE_PAREN;
 * ```
 */
export class ErrorParametersDeclaration {
    /**
     * Constructs a new AST node of type `ErrorParametersDeclaration`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `ErrorParametersDeclaration`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const [$openParen, $parameters, $closeParen] = wasm.ast.Selectors.sequence(this.cst);
            return {
                openParen: $openParen,
                parameters: new ErrorParameters($parameters),
                closeParen: $closeParen,
            };
        });
        assertKind(this.cst.kind, NonterminalKind.ErrorParametersDeclaration);
    }
    /**
     * Returns the child node that has the label `open_paren`.
     */
    get openParen() {
        return this.fetch().openParen;
    }
    /**
     * Returns the child node that has the label `parameters`.
     */
    get parameters() {
        return this.fetch().parameters;
    }
    /**
     * Returns the child node that has the label `close_paren`.
     */
    get closeParen() {
        return this.fetch().closeParen;
    }
}
/**
 * This node represents a `ErrorParameter` nonterminal, with the following structure:
 *
 * ```ebnf
 * (* Introduced in 0.8.4 *)
 * ErrorParameter = (* type_name: *) TypeName
 *                  (* name: *) IDENTIFIER?;
 * ```
 */
export class ErrorParameter {
    /**
     * Constructs a new AST node of type `ErrorParameter`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `ErrorParameter`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const [$typeName, $name] = wasm.ast.Selectors.sequence(this.cst);
            return {
                typeName: new TypeName($typeName),
                name: $name === undefined ? undefined : $name,
            };
        });
        assertKind(this.cst.kind, NonterminalKind.ErrorParameter);
    }
    /**
     * Returns the child node that has the label `type_name`.
     */
    get typeName() {
        return this.fetch().typeName;
    }
    /**
     * Returns the child node that has the label `name`.
     */
    get name() {
        return this.fetch().name;
    }
}
/**
 * This node represents a `ArrayTypeName` nonterminal, with the following structure:
 *
 * ```ebnf
 * (* Postfix unary operator *)
 * ArrayTypeName = (* operand: *) TypeName
 *                 (* open_bracket: *) OPEN_BRACKET
 *                 (* index: *) Expression?
 *                 (* close_bracket: *) CLOSE_BRACKET;
 * ```
 */
export class ArrayTypeName {
    /**
     * Constructs a new AST node of type `ArrayTypeName`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `ArrayTypeName`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const [$operand, $openBracket, $index, $closeBracket] = wasm.ast.Selectors.sequence(this.cst);
            return {
                operand: new TypeName($operand),
                openBracket: $openBracket,
                index: $index === undefined ? undefined : new Expression($index),
                closeBracket: $closeBracket,
            };
        });
        assertKind(this.cst.kind, NonterminalKind.ArrayTypeName);
    }
    /**
     * Returns the child node that has the label `operand`.
     */
    get operand() {
        return this.fetch().operand;
    }
    /**
     * Returns the child node that has the label `open_bracket`.
     */
    get openBracket() {
        return this.fetch().openBracket;
    }
    /**
     * Returns the child node that has the label `index`.
     */
    get index() {
        return this.fetch().index;
    }
    /**
     * Returns the child node that has the label `close_bracket`.
     */
    get closeBracket() {
        return this.fetch().closeBracket;
    }
}
/**
 * This node represents a `FunctionType` nonterminal, with the following structure:
 *
 * ```ebnf
 * FunctionType = (* function_keyword: *) FUNCTION_KEYWORD
 *                (* parameters: *) ParametersDeclaration
 *                (* attributes: *) FunctionTypeAttributes
 *                (* returns: *) ReturnsDeclaration?;
 * ```
 */
export class FunctionType {
    /**
     * Constructs a new AST node of type `FunctionType`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `FunctionType`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const [$functionKeyword, $parameters, $attributes, $returns] = wasm.ast.Selectors.sequence(this.cst);
            return {
                functionKeyword: $functionKeyword,
                parameters: new ParametersDeclaration($parameters),
                attributes: new FunctionTypeAttributes($attributes),
                returns: $returns === undefined ? undefined : new ReturnsDeclaration($returns),
            };
        });
        assertKind(this.cst.kind, NonterminalKind.FunctionType);
    }
    /**
     * Returns the child node that has the label `function_keyword`.
     */
    get functionKeyword() {
        return this.fetch().functionKeyword;
    }
    /**
     * Returns the child node that has the label `parameters`.
     */
    get parameters() {
        return this.fetch().parameters;
    }
    /**
     * Returns the child node that has the label `attributes`.
     */
    get attributes() {
        return this.fetch().attributes;
    }
    /**
     * Returns the child node that has the label `returns`.
     */
    get returns() {
        return this.fetch().returns;
    }
}
/**
 * This node represents a `MappingType` nonterminal, with the following structure:
 *
 * ```ebnf
 * MappingType = (* mapping_keyword: *) MAPPING_KEYWORD
 *               (* open_paren: *) OPEN_PAREN
 *               (* key_type: *) MappingKey
 *               (* equal_greater_than: *) EQUAL_GREATER_THAN
 *               (* value_type: *) MappingValue
 *               (* close_paren: *) CLOSE_PAREN;
 * ```
 */
export class MappingType {
    /**
     * Constructs a new AST node of type `MappingType`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `MappingType`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const [$mappingKeyword, $openParen, $keyType, $equalGreaterThan, $valueType, $closeParen] = wasm.ast.Selectors.sequence(this.cst);
            return {
                mappingKeyword: $mappingKeyword,
                openParen: $openParen,
                keyType: new MappingKey($keyType),
                equalGreaterThan: $equalGreaterThan,
                valueType: new MappingValue($valueType),
                closeParen: $closeParen,
            };
        });
        assertKind(this.cst.kind, NonterminalKind.MappingType);
    }
    /**
     * Returns the child node that has the label `mapping_keyword`.
     */
    get mappingKeyword() {
        return this.fetch().mappingKeyword;
    }
    /**
     * Returns the child node that has the label `open_paren`.
     */
    get openParen() {
        return this.fetch().openParen;
    }
    /**
     * Returns the child node that has the label `key_type`.
     */
    get keyType() {
        return this.fetch().keyType;
    }
    /**
     * Returns the child node that has the label `equal_greater_than`.
     */
    get equalGreaterThan() {
        return this.fetch().equalGreaterThan;
    }
    /**
     * Returns the child node that has the label `value_type`.
     */
    get valueType() {
        return this.fetch().valueType;
    }
    /**
     * Returns the child node that has the label `close_paren`.
     */
    get closeParen() {
        return this.fetch().closeParen;
    }
}
/**
 * This node represents a `MappingKey` nonterminal, with the following structure:
 *
 * ```ebnf
 * MappingKey = (* key_type: *) MappingKeyType
 *              (* name: *) IDENTIFIER?; (* Introduced in 0.8.18 *)
 * ```
 */
export class MappingKey {
    /**
     * Constructs a new AST node of type `MappingKey`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `MappingKey`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const [$keyType, $name] = wasm.ast.Selectors.sequence(this.cst);
            return {
                keyType: new MappingKeyType($keyType),
                name: $name === undefined ? undefined : $name,
            };
        });
        assertKind(this.cst.kind, NonterminalKind.MappingKey);
    }
    /**
     * Returns the child node that has the label `key_type`.
     */
    get keyType() {
        return this.fetch().keyType;
    }
    /**
     * Returns the child node that has the label `name`.
     */
    get name() {
        return this.fetch().name;
    }
}
/**
 * This node represents a `MappingValue` nonterminal, with the following structure:
 *
 * ```ebnf
 * MappingValue = (* type_name: *) TypeName
 *                (* name: *) IDENTIFIER?; (* Introduced in 0.8.18 *)
 * ```
 */
export class MappingValue {
    /**
     * Constructs a new AST node of type `MappingValue`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `MappingValue`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const [$typeName, $name] = wasm.ast.Selectors.sequence(this.cst);
            return {
                typeName: new TypeName($typeName),
                name: $name === undefined ? undefined : $name,
            };
        });
        assertKind(this.cst.kind, NonterminalKind.MappingValue);
    }
    /**
     * Returns the child node that has the label `type_name`.
     */
    get typeName() {
        return this.fetch().typeName;
    }
    /**
     * Returns the child node that has the label `name`.
     */
    get name() {
        return this.fetch().name;
    }
}
/**
 * This node represents a `AddressType` nonterminal, with the following structure:
 *
 * ```ebnf
 * AddressType = (* address_keyword: *) ADDRESS_KEYWORD
 *               (* payable_keyword: *) PAYABLE_KEYWORD?; (* Introduced in 0.5.0 *)
 * ```
 */
export class AddressType {
    /**
     * Constructs a new AST node of type `AddressType`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `AddressType`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const [$addressKeyword, $payableKeyword] = wasm.ast.Selectors.sequence(this.cst);
            return {
                addressKeyword: $addressKeyword,
                payableKeyword: $payableKeyword === undefined ? undefined : $payableKeyword,
            };
        });
        assertKind(this.cst.kind, NonterminalKind.AddressType);
    }
    /**
     * Returns the child node that has the label `address_keyword`.
     */
    get addressKeyword() {
        return this.fetch().addressKeyword;
    }
    /**
     * Returns the child node that has the label `payable_keyword`.
     */
    get payableKeyword() {
        return this.fetch().payableKeyword;
    }
}
/**
 * This node represents a `Block` nonterminal, with the following structure:
 *
 * ```ebnf
 * Block = (* open_brace: *) OPEN_BRACE
 *         (* statements: *) Statements
 *         (* close_brace: *) CLOSE_BRACE;
 * ```
 */
export class Block {
    /**
     * Constructs a new AST node of type `Block`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `Block`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const [$openBrace, $statements, $closeBrace] = wasm.ast.Selectors.sequence(this.cst);
            return {
                openBrace: $openBrace,
                statements: new Statements($statements),
                closeBrace: $closeBrace,
            };
        });
        assertKind(this.cst.kind, NonterminalKind.Block);
    }
    /**
     * Returns the child node that has the label `open_brace`.
     */
    get openBrace() {
        return this.fetch().openBrace;
    }
    /**
     * Returns the child node that has the label `statements`.
     */
    get statements() {
        return this.fetch().statements;
    }
    /**
     * Returns the child node that has the label `close_brace`.
     */
    get closeBrace() {
        return this.fetch().closeBrace;
    }
}
/**
 * This node represents a `UncheckedBlock` nonterminal, with the following structure:
 *
 * ```ebnf
 * (* Introduced in 0.8.0 *)
 * UncheckedBlock = (* unchecked_keyword: *) UNCHECKED_KEYWORD
 *                  (* block: *) Block;
 * ```
 */
export class UncheckedBlock {
    /**
     * Constructs a new AST node of type `UncheckedBlock`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `UncheckedBlock`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const [$uncheckedKeyword, $block] = wasm.ast.Selectors.sequence(this.cst);
            return {
                uncheckedKeyword: $uncheckedKeyword,
                block: new Block($block),
            };
        });
        assertKind(this.cst.kind, NonterminalKind.UncheckedBlock);
    }
    /**
     * Returns the child node that has the label `unchecked_keyword`.
     */
    get uncheckedKeyword() {
        return this.fetch().uncheckedKeyword;
    }
    /**
     * Returns the child node that has the label `block`.
     */
    get block() {
        return this.fetch().block;
    }
}
/**
 * This node represents a `ExpressionStatement` nonterminal, with the following structure:
 *
 * ```ebnf
 * ExpressionStatement = (* expression: *) Expression
 *                       (* semicolon: *) SEMICOLON;
 * ```
 */
export class ExpressionStatement {
    /**
     * Constructs a new AST node of type `ExpressionStatement`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `ExpressionStatement`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const [$expression, $semicolon] = wasm.ast.Selectors.sequence(this.cst);
            return {
                expression: new Expression($expression),
                semicolon: $semicolon,
            };
        });
        assertKind(this.cst.kind, NonterminalKind.ExpressionStatement);
    }
    /**
     * Returns the child node that has the label `expression`.
     */
    get expression() {
        return this.fetch().expression;
    }
    /**
     * Returns the child node that has the label `semicolon`.
     */
    get semicolon() {
        return this.fetch().semicolon;
    }
}
/**
 * This node represents a `AssemblyStatement` nonterminal, with the following structure:
 *
 * ```ebnf
 * AssemblyStatement = (* assembly_keyword: *) ASSEMBLY_KEYWORD
 *                     (* label: *) StringLiteral?
 *                     (* flags: *) AssemblyFlagsDeclaration? (* Introduced in 0.8.13 *)
 *                     (* body: *) YulBlock;
 * ```
 */
export class AssemblyStatement {
    /**
     * Constructs a new AST node of type `AssemblyStatement`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `AssemblyStatement`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const [$assemblyKeyword, $label, $flags, $body] = wasm.ast.Selectors.sequence(this.cst);
            return {
                assemblyKeyword: $assemblyKeyword,
                label: $label === undefined ? undefined : new StringLiteral($label),
                flags: $flags === undefined ? undefined : new AssemblyFlagsDeclaration($flags),
                body: new YulBlock($body),
            };
        });
        assertKind(this.cst.kind, NonterminalKind.AssemblyStatement);
    }
    /**
     * Returns the child node that has the label `assembly_keyword`.
     */
    get assemblyKeyword() {
        return this.fetch().assemblyKeyword;
    }
    /**
     * Returns the child node that has the label `label`.
     */
    get label() {
        return this.fetch().label;
    }
    /**
     * Returns the child node that has the label `flags`.
     */
    get flags() {
        return this.fetch().flags;
    }
    /**
     * Returns the child node that has the label `body`.
     */
    get body() {
        return this.fetch().body;
    }
}
/**
 * This node represents a `AssemblyFlagsDeclaration` nonterminal, with the following structure:
 *
 * ```ebnf
 * (* Introduced in 0.8.13 *)
 * AssemblyFlagsDeclaration = (* open_paren: *) OPEN_PAREN
 *                            (* flags: *) AssemblyFlags
 *                            (* close_paren: *) CLOSE_PAREN;
 * ```
 */
export class AssemblyFlagsDeclaration {
    /**
     * Constructs a new AST node of type `AssemblyFlagsDeclaration`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `AssemblyFlagsDeclaration`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const [$openParen, $flags, $closeParen] = wasm.ast.Selectors.sequence(this.cst);
            return {
                openParen: $openParen,
                flags: new AssemblyFlags($flags),
                closeParen: $closeParen,
            };
        });
        assertKind(this.cst.kind, NonterminalKind.AssemblyFlagsDeclaration);
    }
    /**
     * Returns the child node that has the label `open_paren`.
     */
    get openParen() {
        return this.fetch().openParen;
    }
    /**
     * Returns the child node that has the label `flags`.
     */
    get flags() {
        return this.fetch().flags;
    }
    /**
     * Returns the child node that has the label `close_paren`.
     */
    get closeParen() {
        return this.fetch().closeParen;
    }
}
/**
 * This node represents a `TupleDeconstructionStatement` nonterminal, with the following structure:
 *
 * ```ebnf
 * TupleDeconstructionStatement = (* var_keyword: *) VAR_KEYWORD? (* Deprecated in 0.5.0 *)
 *                                (* open_paren: *) OPEN_PAREN
 *                                (* elements: *) TupleDeconstructionElements
 *                                (* close_paren: *) CLOSE_PAREN
 *                                (* equal: *) EQUAL
 *                                (* expression: *) Expression
 *                                (* semicolon: *) SEMICOLON;
 * ```
 */
export class TupleDeconstructionStatement {
    /**
     * Constructs a new AST node of type `TupleDeconstructionStatement`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `TupleDeconstructionStatement`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const [$varKeyword, $openParen, $elements, $closeParen, $equal, $expression, $semicolon] = wasm.ast.Selectors.sequence(this.cst);
            return {
                varKeyword: $varKeyword === undefined ? undefined : $varKeyword,
                openParen: $openParen,
                elements: new TupleDeconstructionElements($elements),
                closeParen: $closeParen,
                equal: $equal,
                expression: new Expression($expression),
                semicolon: $semicolon,
            };
        });
        assertKind(this.cst.kind, NonterminalKind.TupleDeconstructionStatement);
    }
    /**
     * Returns the child node that has the label `var_keyword`.
     */
    get varKeyword() {
        return this.fetch().varKeyword;
    }
    /**
     * Returns the child node that has the label `open_paren`.
     */
    get openParen() {
        return this.fetch().openParen;
    }
    /**
     * Returns the child node that has the label `elements`.
     */
    get elements() {
        return this.fetch().elements;
    }
    /**
     * Returns the child node that has the label `close_paren`.
     */
    get closeParen() {
        return this.fetch().closeParen;
    }
    /**
     * Returns the child node that has the label `equal`.
     */
    get equal() {
        return this.fetch().equal;
    }
    /**
     * Returns the child node that has the label `expression`.
     */
    get expression() {
        return this.fetch().expression;
    }
    /**
     * Returns the child node that has the label `semicolon`.
     */
    get semicolon() {
        return this.fetch().semicolon;
    }
}
/**
 * This node represents a `TupleDeconstructionElement` nonterminal, with the following structure:
 *
 * ```ebnf
 * TupleDeconstructionElement = (* member: *) TupleMember?;
 * ```
 */
export class TupleDeconstructionElement {
    /**
     * Constructs a new AST node of type `TupleDeconstructionElement`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `TupleDeconstructionElement`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const [$member] = wasm.ast.Selectors.sequence(this.cst);
            return {
                member: $member === undefined ? undefined : new TupleMember($member),
            };
        });
        assertKind(this.cst.kind, NonterminalKind.TupleDeconstructionElement);
    }
    /**
     * Returns the child node that has the label `member`.
     */
    get member() {
        return this.fetch().member;
    }
}
/**
 * This node represents a `TypedTupleMember` nonterminal, with the following structure:
 *
 * ```ebnf
 * TypedTupleMember = (* type_name: *) TypeName
 *                    (* storage_location: *) StorageLocation?
 *                    (* name: *) IDENTIFIER;
 * ```
 */
export class TypedTupleMember {
    /**
     * Constructs a new AST node of type `TypedTupleMember`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `TypedTupleMember`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const [$typeName, $storageLocation, $name] = wasm.ast.Selectors.sequence(this.cst);
            return {
                typeName: new TypeName($typeName),
                storageLocation: $storageLocation === undefined ? undefined : new StorageLocation($storageLocation),
                name: $name,
            };
        });
        assertKind(this.cst.kind, NonterminalKind.TypedTupleMember);
    }
    /**
     * Returns the child node that has the label `type_name`.
     */
    get typeName() {
        return this.fetch().typeName;
    }
    /**
     * Returns the child node that has the label `storage_location`.
     */
    get storageLocation() {
        return this.fetch().storageLocation;
    }
    /**
     * Returns the child node that has the label `name`.
     */
    get name() {
        return this.fetch().name;
    }
}
/**
 * This node represents a `UntypedTupleMember` nonterminal, with the following structure:
 *
 * ```ebnf
 * UntypedTupleMember = (* storage_location: *) StorageLocation?
 *                      (* name: *) IDENTIFIER;
 * ```
 */
export class UntypedTupleMember {
    /**
     * Constructs a new AST node of type `UntypedTupleMember`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `UntypedTupleMember`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const [$storageLocation, $name] = wasm.ast.Selectors.sequence(this.cst);
            return {
                storageLocation: $storageLocation === undefined ? undefined : new StorageLocation($storageLocation),
                name: $name,
            };
        });
        assertKind(this.cst.kind, NonterminalKind.UntypedTupleMember);
    }
    /**
     * Returns the child node that has the label `storage_location`.
     */
    get storageLocation() {
        return this.fetch().storageLocation;
    }
    /**
     * Returns the child node that has the label `name`.
     */
    get name() {
        return this.fetch().name;
    }
}
/**
 * This node represents a `VariableDeclarationStatement` nonterminal, with the following structure:
 *
 * ```ebnf
 * VariableDeclarationStatement = (* variable_type: *) VariableDeclarationType
 *                                (* storage_location: *) StorageLocation?
 *                                (* name: *) IDENTIFIER
 *                                (* value: *) VariableDeclarationValue?
 *                                (* semicolon: *) SEMICOLON;
 * ```
 */
export class VariableDeclarationStatement {
    /**
     * Constructs a new AST node of type `VariableDeclarationStatement`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `VariableDeclarationStatement`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const [$variableType, $storageLocation, $name, $value, $semicolon] = wasm.ast.Selectors.sequence(this.cst);
            return {
                variableType: new VariableDeclarationType($variableType),
                storageLocation: $storageLocation === undefined ? undefined : new StorageLocation($storageLocation),
                name: $name,
                value: $value === undefined ? undefined : new VariableDeclarationValue($value),
                semicolon: $semicolon,
            };
        });
        assertKind(this.cst.kind, NonterminalKind.VariableDeclarationStatement);
    }
    /**
     * Returns the child node that has the label `variable_type`.
     */
    get variableType() {
        return this.fetch().variableType;
    }
    /**
     * Returns the child node that has the label `storage_location`.
     */
    get storageLocation() {
        return this.fetch().storageLocation;
    }
    /**
     * Returns the child node that has the label `name`.
     */
    get name() {
        return this.fetch().name;
    }
    /**
     * Returns the child node that has the label `value`.
     */
    get value() {
        return this.fetch().value;
    }
    /**
     * Returns the child node that has the label `semicolon`.
     */
    get semicolon() {
        return this.fetch().semicolon;
    }
}
/**
 * This node represents a `VariableDeclarationValue` nonterminal, with the following structure:
 *
 * ```ebnf
 * VariableDeclarationValue = (* equal: *) EQUAL
 *                            (* expression: *) Expression;
 * ```
 */
export class VariableDeclarationValue {
    /**
     * Constructs a new AST node of type `VariableDeclarationValue`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `VariableDeclarationValue`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const [$equal, $expression] = wasm.ast.Selectors.sequence(this.cst);
            return {
                equal: $equal,
                expression: new Expression($expression),
            };
        });
        assertKind(this.cst.kind, NonterminalKind.VariableDeclarationValue);
    }
    /**
     * Returns the child node that has the label `equal`.
     */
    get equal() {
        return this.fetch().equal;
    }
    /**
     * Returns the child node that has the label `expression`.
     */
    get expression() {
        return this.fetch().expression;
    }
}
/**
 * This node represents a `IfStatement` nonterminal, with the following structure:
 *
 * ```ebnf
 * IfStatement = (* if_keyword: *) IF_KEYWORD
 *               (* open_paren: *) OPEN_PAREN
 *               (* condition: *) Expression
 *               (* close_paren: *) CLOSE_PAREN
 *               (* body: *) Statement
 *               (* else_branch: *) ElseBranch?;
 * ```
 */
export class IfStatement {
    /**
     * Constructs a new AST node of type `IfStatement`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `IfStatement`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const [$ifKeyword, $openParen, $condition, $closeParen, $body, $elseBranch] = wasm.ast.Selectors.sequence(this.cst);
            return {
                ifKeyword: $ifKeyword,
                openParen: $openParen,
                condition: new Expression($condition),
                closeParen: $closeParen,
                body: new Statement($body),
                elseBranch: $elseBranch === undefined ? undefined : new ElseBranch($elseBranch),
            };
        });
        assertKind(this.cst.kind, NonterminalKind.IfStatement);
    }
    /**
     * Returns the child node that has the label `if_keyword`.
     */
    get ifKeyword() {
        return this.fetch().ifKeyword;
    }
    /**
     * Returns the child node that has the label `open_paren`.
     */
    get openParen() {
        return this.fetch().openParen;
    }
    /**
     * Returns the child node that has the label `condition`.
     */
    get condition() {
        return this.fetch().condition;
    }
    /**
     * Returns the child node that has the label `close_paren`.
     */
    get closeParen() {
        return this.fetch().closeParen;
    }
    /**
     * Returns the child node that has the label `body`.
     */
    get body() {
        return this.fetch().body;
    }
    /**
     * Returns the child node that has the label `else_branch`.
     */
    get elseBranch() {
        return this.fetch().elseBranch;
    }
}
/**
 * This node represents a `ElseBranch` nonterminal, with the following structure:
 *
 * ```ebnf
 * ElseBranch = (* else_keyword: *) ELSE_KEYWORD
 *              (* body: *) Statement;
 * ```
 */
export class ElseBranch {
    /**
     * Constructs a new AST node of type `ElseBranch`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `ElseBranch`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const [$elseKeyword, $body] = wasm.ast.Selectors.sequence(this.cst);
            return {
                elseKeyword: $elseKeyword,
                body: new Statement($body),
            };
        });
        assertKind(this.cst.kind, NonterminalKind.ElseBranch);
    }
    /**
     * Returns the child node that has the label `else_keyword`.
     */
    get elseKeyword() {
        return this.fetch().elseKeyword;
    }
    /**
     * Returns the child node that has the label `body`.
     */
    get body() {
        return this.fetch().body;
    }
}
/**
 * This node represents a `ForStatement` nonterminal, with the following structure:
 *
 * ```ebnf
 * ForStatement = (* for_keyword: *) FOR_KEYWORD
 *                (* open_paren: *) OPEN_PAREN
 *                (* initialization: *) ForStatementInitialization
 *                (* condition: *) ForStatementCondition
 *                (* iterator: *) Expression?
 *                (* close_paren: *) CLOSE_PAREN
 *                (* body: *) Statement;
 * ```
 */
export class ForStatement {
    /**
     * Constructs a new AST node of type `ForStatement`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `ForStatement`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const [$forKeyword, $openParen, $initialization, $condition, $iterator, $closeParen, $body] = wasm.ast.Selectors.sequence(this.cst);
            return {
                forKeyword: $forKeyword,
                openParen: $openParen,
                initialization: new ForStatementInitialization($initialization),
                condition: new ForStatementCondition($condition),
                iterator: $iterator === undefined ? undefined : new Expression($iterator),
                closeParen: $closeParen,
                body: new Statement($body),
            };
        });
        assertKind(this.cst.kind, NonterminalKind.ForStatement);
    }
    /**
     * Returns the child node that has the label `for_keyword`.
     */
    get forKeyword() {
        return this.fetch().forKeyword;
    }
    /**
     * Returns the child node that has the label `open_paren`.
     */
    get openParen() {
        return this.fetch().openParen;
    }
    /**
     * Returns the child node that has the label `initialization`.
     */
    get initialization() {
        return this.fetch().initialization;
    }
    /**
     * Returns the child node that has the label `condition`.
     */
    get condition() {
        return this.fetch().condition;
    }
    /**
     * Returns the child node that has the label `iterator`.
     */
    get iterator() {
        return this.fetch().iterator;
    }
    /**
     * Returns the child node that has the label `close_paren`.
     */
    get closeParen() {
        return this.fetch().closeParen;
    }
    /**
     * Returns the child node that has the label `body`.
     */
    get body() {
        return this.fetch().body;
    }
}
/**
 * This node represents a `WhileStatement` nonterminal, with the following structure:
 *
 * ```ebnf
 * WhileStatement = (* while_keyword: *) WHILE_KEYWORD
 *                  (* open_paren: *) OPEN_PAREN
 *                  (* condition: *) Expression
 *                  (* close_paren: *) CLOSE_PAREN
 *                  (* body: *) Statement;
 * ```
 */
export class WhileStatement {
    /**
     * Constructs a new AST node of type `WhileStatement`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `WhileStatement`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const [$whileKeyword, $openParen, $condition, $closeParen, $body] = wasm.ast.Selectors.sequence(this.cst);
            return {
                whileKeyword: $whileKeyword,
                openParen: $openParen,
                condition: new Expression($condition),
                closeParen: $closeParen,
                body: new Statement($body),
            };
        });
        assertKind(this.cst.kind, NonterminalKind.WhileStatement);
    }
    /**
     * Returns the child node that has the label `while_keyword`.
     */
    get whileKeyword() {
        return this.fetch().whileKeyword;
    }
    /**
     * Returns the child node that has the label `open_paren`.
     */
    get openParen() {
        return this.fetch().openParen;
    }
    /**
     * Returns the child node that has the label `condition`.
     */
    get condition() {
        return this.fetch().condition;
    }
    /**
     * Returns the child node that has the label `close_paren`.
     */
    get closeParen() {
        return this.fetch().closeParen;
    }
    /**
     * Returns the child node that has the label `body`.
     */
    get body() {
        return this.fetch().body;
    }
}
/**
 * This node represents a `DoWhileStatement` nonterminal, with the following structure:
 *
 * ```ebnf
 * DoWhileStatement = (* do_keyword: *) DO_KEYWORD
 *                    (* body: *) Statement
 *                    (* while_keyword: *) WHILE_KEYWORD
 *                    (* open_paren: *) OPEN_PAREN
 *                    (* condition: *) Expression
 *                    (* close_paren: *) CLOSE_PAREN
 *                    (* semicolon: *) SEMICOLON;
 * ```
 */
export class DoWhileStatement {
    /**
     * Constructs a new AST node of type `DoWhileStatement`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `DoWhileStatement`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const [$doKeyword, $body, $whileKeyword, $openParen, $condition, $closeParen, $semicolon] = wasm.ast.Selectors.sequence(this.cst);
            return {
                doKeyword: $doKeyword,
                body: new Statement($body),
                whileKeyword: $whileKeyword,
                openParen: $openParen,
                condition: new Expression($condition),
                closeParen: $closeParen,
                semicolon: $semicolon,
            };
        });
        assertKind(this.cst.kind, NonterminalKind.DoWhileStatement);
    }
    /**
     * Returns the child node that has the label `do_keyword`.
     */
    get doKeyword() {
        return this.fetch().doKeyword;
    }
    /**
     * Returns the child node that has the label `body`.
     */
    get body() {
        return this.fetch().body;
    }
    /**
     * Returns the child node that has the label `while_keyword`.
     */
    get whileKeyword() {
        return this.fetch().whileKeyword;
    }
    /**
     * Returns the child node that has the label `open_paren`.
     */
    get openParen() {
        return this.fetch().openParen;
    }
    /**
     * Returns the child node that has the label `condition`.
     */
    get condition() {
        return this.fetch().condition;
    }
    /**
     * Returns the child node that has the label `close_paren`.
     */
    get closeParen() {
        return this.fetch().closeParen;
    }
    /**
     * Returns the child node that has the label `semicolon`.
     */
    get semicolon() {
        return this.fetch().semicolon;
    }
}
/**
 * This node represents a `ContinueStatement` nonterminal, with the following structure:
 *
 * ```ebnf
 * ContinueStatement = (* continue_keyword: *) CONTINUE_KEYWORD
 *                     (* semicolon: *) SEMICOLON;
 * ```
 */
export class ContinueStatement {
    /**
     * Constructs a new AST node of type `ContinueStatement`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `ContinueStatement`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const [$continueKeyword, $semicolon] = wasm.ast.Selectors.sequence(this.cst);
            return {
                continueKeyword: $continueKeyword,
                semicolon: $semicolon,
            };
        });
        assertKind(this.cst.kind, NonterminalKind.ContinueStatement);
    }
    /**
     * Returns the child node that has the label `continue_keyword`.
     */
    get continueKeyword() {
        return this.fetch().continueKeyword;
    }
    /**
     * Returns the child node that has the label `semicolon`.
     */
    get semicolon() {
        return this.fetch().semicolon;
    }
}
/**
 * This node represents a `BreakStatement` nonterminal, with the following structure:
 *
 * ```ebnf
 * BreakStatement = (* break_keyword: *) BREAK_KEYWORD
 *                  (* semicolon: *) SEMICOLON;
 * ```
 */
export class BreakStatement {
    /**
     * Constructs a new AST node of type `BreakStatement`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `BreakStatement`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const [$breakKeyword, $semicolon] = wasm.ast.Selectors.sequence(this.cst);
            return {
                breakKeyword: $breakKeyword,
                semicolon: $semicolon,
            };
        });
        assertKind(this.cst.kind, NonterminalKind.BreakStatement);
    }
    /**
     * Returns the child node that has the label `break_keyword`.
     */
    get breakKeyword() {
        return this.fetch().breakKeyword;
    }
    /**
     * Returns the child node that has the label `semicolon`.
     */
    get semicolon() {
        return this.fetch().semicolon;
    }
}
/**
 * This node represents a `ReturnStatement` nonterminal, with the following structure:
 *
 * ```ebnf
 * ReturnStatement = (* return_keyword: *) RETURN_KEYWORD
 *                   (* expression: *) Expression?
 *                   (* semicolon: *) SEMICOLON;
 * ```
 */
export class ReturnStatement {
    /**
     * Constructs a new AST node of type `ReturnStatement`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `ReturnStatement`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const [$returnKeyword, $expression, $semicolon] = wasm.ast.Selectors.sequence(this.cst);
            return {
                returnKeyword: $returnKeyword,
                expression: $expression === undefined ? undefined : new Expression($expression),
                semicolon: $semicolon,
            };
        });
        assertKind(this.cst.kind, NonterminalKind.ReturnStatement);
    }
    /**
     * Returns the child node that has the label `return_keyword`.
     */
    get returnKeyword() {
        return this.fetch().returnKeyword;
    }
    /**
     * Returns the child node that has the label `expression`.
     */
    get expression() {
        return this.fetch().expression;
    }
    /**
     * Returns the child node that has the label `semicolon`.
     */
    get semicolon() {
        return this.fetch().semicolon;
    }
}
/**
 * This node represents a `EmitStatement` nonterminal, with the following structure:
 *
 * ```ebnf
 * (* Introduced in 0.4.21 *)
 * EmitStatement = (* emit_keyword: *) EMIT_KEYWORD
 *                 (* event: *) IdentifierPath
 *                 (* arguments: *) ArgumentsDeclaration
 *                 (* semicolon: *) SEMICOLON;
 * ```
 */
export class EmitStatement {
    /**
     * Constructs a new AST node of type `EmitStatement`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `EmitStatement`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const [$emitKeyword, $event, $arguments, $semicolon] = wasm.ast.Selectors.sequence(this.cst);
            return {
                emitKeyword: $emitKeyword,
                event: new IdentifierPath($event),
                arguments: new ArgumentsDeclaration($arguments),
                semicolon: $semicolon,
            };
        });
        assertKind(this.cst.kind, NonterminalKind.EmitStatement);
    }
    /**
     * Returns the child node that has the label `emit_keyword`.
     */
    get emitKeyword() {
        return this.fetch().emitKeyword;
    }
    /**
     * Returns the child node that has the label `event`.
     */
    get event() {
        return this.fetch().event;
    }
    /**
     * Returns the child node that has the label `arguments`.
     */
    get arguments() {
        return this.fetch().arguments;
    }
    /**
     * Returns the child node that has the label `semicolon`.
     */
    get semicolon() {
        return this.fetch().semicolon;
    }
}
/**
 * This node represents a `TryStatement` nonterminal, with the following structure:
 *
 * ```ebnf
 * (* Introduced in 0.6.0 *)
 * TryStatement = (* try_keyword: *) TRY_KEYWORD
 *                (* expression: *) Expression
 *                (* returns: *) ReturnsDeclaration?
 *                (* body: *) Block
 *                (* catch_clauses: *) CatchClauses;
 * ```
 */
export class TryStatement {
    /**
     * Constructs a new AST node of type `TryStatement`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `TryStatement`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const [$tryKeyword, $expression, $returns, $body, $catchClauses] = wasm.ast.Selectors.sequence(this.cst);
            return {
                tryKeyword: $tryKeyword,
                expression: new Expression($expression),
                returns: $returns === undefined ? undefined : new ReturnsDeclaration($returns),
                body: new Block($body),
                catchClauses: new CatchClauses($catchClauses),
            };
        });
        assertKind(this.cst.kind, NonterminalKind.TryStatement);
    }
    /**
     * Returns the child node that has the label `try_keyword`.
     */
    get tryKeyword() {
        return this.fetch().tryKeyword;
    }
    /**
     * Returns the child node that has the label `expression`.
     */
    get expression() {
        return this.fetch().expression;
    }
    /**
     * Returns the child node that has the label `returns`.
     */
    get returns() {
        return this.fetch().returns;
    }
    /**
     * Returns the child node that has the label `body`.
     */
    get body() {
        return this.fetch().body;
    }
    /**
     * Returns the child node that has the label `catch_clauses`.
     */
    get catchClauses() {
        return this.fetch().catchClauses;
    }
}
/**
 * This node represents a `CatchClause` nonterminal, with the following structure:
 *
 * ```ebnf
 * (* Introduced in 0.6.0 *)
 * CatchClause = (* catch_keyword: *) CATCH_KEYWORD
 *               (* error: *) CatchClauseError?
 *               (* body: *) Block;
 * ```
 */
export class CatchClause {
    /**
     * Constructs a new AST node of type `CatchClause`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `CatchClause`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const [$catchKeyword, $error, $body] = wasm.ast.Selectors.sequence(this.cst);
            return {
                catchKeyword: $catchKeyword,
                error: $error === undefined ? undefined : new CatchClauseError($error),
                body: new Block($body),
            };
        });
        assertKind(this.cst.kind, NonterminalKind.CatchClause);
    }
    /**
     * Returns the child node that has the label `catch_keyword`.
     */
    get catchKeyword() {
        return this.fetch().catchKeyword;
    }
    /**
     * Returns the child node that has the label `error`.
     */
    get error() {
        return this.fetch().error;
    }
    /**
     * Returns the child node that has the label `body`.
     */
    get body() {
        return this.fetch().body;
    }
}
/**
 * This node represents a `CatchClauseError` nonterminal, with the following structure:
 *
 * ```ebnf
 * (* Introduced in 0.6.0 *)
 * CatchClauseError = (* name: *) IDENTIFIER?
 *                    (* parameters: *) ParametersDeclaration;
 * ```
 */
export class CatchClauseError {
    /**
     * Constructs a new AST node of type `CatchClauseError`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `CatchClauseError`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const [$name, $parameters] = wasm.ast.Selectors.sequence(this.cst);
            return {
                name: $name === undefined ? undefined : $name,
                parameters: new ParametersDeclaration($parameters),
            };
        });
        assertKind(this.cst.kind, NonterminalKind.CatchClauseError);
    }
    /**
     * Returns the child node that has the label `name`.
     */
    get name() {
        return this.fetch().name;
    }
    /**
     * Returns the child node that has the label `parameters`.
     */
    get parameters() {
        return this.fetch().parameters;
    }
}
/**
 * This node represents a `RevertStatement` nonterminal, with the following structure:
 *
 * ```ebnf
 * (* Introduced in 0.8.4 *)
 * RevertStatement = (* revert_keyword: *) REVERT_KEYWORD
 *                   (* error: *) IdentifierPath
 *                   (* arguments: *) ArgumentsDeclaration
 *                   (* semicolon: *) SEMICOLON;
 * ```
 */
export class RevertStatement {
    /**
     * Constructs a new AST node of type `RevertStatement`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `RevertStatement`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const [$revertKeyword, $error, $arguments, $semicolon] = wasm.ast.Selectors.sequence(this.cst);
            return {
                revertKeyword: $revertKeyword,
                error: new IdentifierPath($error),
                arguments: new ArgumentsDeclaration($arguments),
                semicolon: $semicolon,
            };
        });
        assertKind(this.cst.kind, NonterminalKind.RevertStatement);
    }
    /**
     * Returns the child node that has the label `revert_keyword`.
     */
    get revertKeyword() {
        return this.fetch().revertKeyword;
    }
    /**
     * Returns the child node that has the label `error`.
     */
    get error() {
        return this.fetch().error;
    }
    /**
     * Returns the child node that has the label `arguments`.
     */
    get arguments() {
        return this.fetch().arguments;
    }
    /**
     * Returns the child node that has the label `semicolon`.
     */
    get semicolon() {
        return this.fetch().semicolon;
    }
}
/**
 * This node represents a `ThrowStatement` nonterminal, with the following structure:
 *
 * ```ebnf
 * (* Deprecated in 0.5.0 *)
 * ThrowStatement = (* throw_keyword: *) THROW_KEYWORD
 *                  (* semicolon: *) SEMICOLON;
 * ```
 */
export class ThrowStatement {
    /**
     * Constructs a new AST node of type `ThrowStatement`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `ThrowStatement`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const [$throwKeyword, $semicolon] = wasm.ast.Selectors.sequence(this.cst);
            return {
                throwKeyword: $throwKeyword,
                semicolon: $semicolon,
            };
        });
        assertKind(this.cst.kind, NonterminalKind.ThrowStatement);
    }
    /**
     * Returns the child node that has the label `throw_keyword`.
     */
    get throwKeyword() {
        return this.fetch().throwKeyword;
    }
    /**
     * Returns the child node that has the label `semicolon`.
     */
    get semicolon() {
        return this.fetch().semicolon;
    }
}
/**
 * This node represents a `AssignmentExpression` nonterminal, with the following structure:
 *
 * ```ebnf
 * (* Right-associative binary operator *)
 * AssignmentExpression = (* left_operand: *) Expression
 *                        (* operator: *) EQUAL
 *                        (* right_operand: *) Expression;
 *
 * (* Right-associative binary operator *)
 * AssignmentExpression = (* left_operand: *) Expression
 *                        (* operator: *) BAR_EQUAL
 *                        (* right_operand: *) Expression;
 *
 * (* Right-associative binary operator *)
 * AssignmentExpression = (* left_operand: *) Expression
 *                        (* operator: *) PLUS_EQUAL
 *                        (* right_operand: *) Expression;
 *
 * (* Right-associative binary operator *)
 * AssignmentExpression = (* left_operand: *) Expression
 *                        (* operator: *) MINUS_EQUAL
 *                        (* right_operand: *) Expression;
 *
 * (* Right-associative binary operator *)
 * AssignmentExpression = (* left_operand: *) Expression
 *                        (* operator: *) CARET_EQUAL
 *                        (* right_operand: *) Expression;
 *
 * (* Right-associative binary operator *)
 * AssignmentExpression = (* left_operand: *) Expression
 *                        (* operator: *) SLASH_EQUAL
 *                        (* right_operand: *) Expression;
 *
 * (* Right-associative binary operator *)
 * AssignmentExpression = (* left_operand: *) Expression
 *                        (* operator: *) PERCENT_EQUAL
 *                        (* right_operand: *) Expression;
 *
 * (* Right-associative binary operator *)
 * AssignmentExpression = (* left_operand: *) Expression
 *                        (* operator: *) ASTERISK_EQUAL
 *                        (* right_operand: *) Expression;
 *
 * (* Right-associative binary operator *)
 * AssignmentExpression = (* left_operand: *) Expression
 *                        (* operator: *) AMPERSAND_EQUAL
 *                        (* right_operand: *) Expression;
 *
 * (* Right-associative binary operator *)
 * AssignmentExpression = (* left_operand: *) Expression
 *                        (* operator: *) LESS_THAN_LESS_THAN_EQUAL
 *                        (* right_operand: *) Expression;
 *
 * (* Right-associative binary operator *)
 * AssignmentExpression = (* left_operand: *) Expression
 *                        (* operator: *) GREATER_THAN_GREATER_THAN_EQUAL
 *                        (* right_operand: *) Expression;
 *
 * (* Right-associative binary operator *)
 * AssignmentExpression = (* left_operand: *) Expression
 *                        (* operator: *) GREATER_THAN_GREATER_THAN_GREATER_THAN_EQUAL
 *                        (* right_operand: *) Expression;
 * ```
 */
export class AssignmentExpression {
    /**
     * Constructs a new AST node of type `AssignmentExpression`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `AssignmentExpression`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const [$leftOperand, $operator, $rightOperand] = wasm.ast.Selectors.sequence(this.cst);
            return {
                leftOperand: new Expression($leftOperand),
                operator: $operator,
                rightOperand: new Expression($rightOperand),
            };
        });
        assertKind(this.cst.kind, NonterminalKind.AssignmentExpression);
    }
    /**
     * Returns the child node that has the label `left_operand`.
     */
    get leftOperand() {
        return this.fetch().leftOperand;
    }
    /**
     * Returns the child node that has the label `operator`.
     */
    get operator() {
        return this.fetch().operator;
    }
    /**
     * Returns the child node that has the label `right_operand`.
     */
    get rightOperand() {
        return this.fetch().rightOperand;
    }
}
/**
 * This node represents a `ConditionalExpression` nonterminal, with the following structure:
 *
 * ```ebnf
 * (* Postfix unary operator *)
 * ConditionalExpression = (* operand: *) Expression
 *                         (* question_mark: *) QUESTION_MARK
 *                         (* true_expression: *) Expression
 *                         (* colon: *) COLON
 *                         (* false_expression: *) Expression;
 * ```
 */
export class ConditionalExpression {
    /**
     * Constructs a new AST node of type `ConditionalExpression`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `ConditionalExpression`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const [$operand, $questionMark, $trueExpression, $colon, $falseExpression] = wasm.ast.Selectors.sequence(this.cst);
            return {
                operand: new Expression($operand),
                questionMark: $questionMark,
                trueExpression: new Expression($trueExpression),
                colon: $colon,
                falseExpression: new Expression($falseExpression),
            };
        });
        assertKind(this.cst.kind, NonterminalKind.ConditionalExpression);
    }
    /**
     * Returns the child node that has the label `operand`.
     */
    get operand() {
        return this.fetch().operand;
    }
    /**
     * Returns the child node that has the label `question_mark`.
     */
    get questionMark() {
        return this.fetch().questionMark;
    }
    /**
     * Returns the child node that has the label `true_expression`.
     */
    get trueExpression() {
        return this.fetch().trueExpression;
    }
    /**
     * Returns the child node that has the label `colon`.
     */
    get colon() {
        return this.fetch().colon;
    }
    /**
     * Returns the child node that has the label `false_expression`.
     */
    get falseExpression() {
        return this.fetch().falseExpression;
    }
}
/**
 * This node represents a `OrExpression` nonterminal, with the following structure:
 *
 * ```ebnf
 * (* Left-associative binary operator *)
 * OrExpression = (* left_operand: *) Expression
 *                (* operator: *) BAR_BAR
 *                (* right_operand: *) Expression;
 * ```
 */
export class OrExpression {
    /**
     * Constructs a new AST node of type `OrExpression`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `OrExpression`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const [$leftOperand, $operator, $rightOperand] = wasm.ast.Selectors.sequence(this.cst);
            return {
                leftOperand: new Expression($leftOperand),
                operator: $operator,
                rightOperand: new Expression($rightOperand),
            };
        });
        assertKind(this.cst.kind, NonterminalKind.OrExpression);
    }
    /**
     * Returns the child node that has the label `left_operand`.
     */
    get leftOperand() {
        return this.fetch().leftOperand;
    }
    /**
     * Returns the child node that has the label `operator`.
     */
    get operator() {
        return this.fetch().operator;
    }
    /**
     * Returns the child node that has the label `right_operand`.
     */
    get rightOperand() {
        return this.fetch().rightOperand;
    }
}
/**
 * This node represents a `AndExpression` nonterminal, with the following structure:
 *
 * ```ebnf
 * (* Left-associative binary operator *)
 * AndExpression = (* left_operand: *) Expression
 *                 (* operator: *) AMPERSAND_AMPERSAND
 *                 (* right_operand: *) Expression;
 * ```
 */
export class AndExpression {
    /**
     * Constructs a new AST node of type `AndExpression`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `AndExpression`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const [$leftOperand, $operator, $rightOperand] = wasm.ast.Selectors.sequence(this.cst);
            return {
                leftOperand: new Expression($leftOperand),
                operator: $operator,
                rightOperand: new Expression($rightOperand),
            };
        });
        assertKind(this.cst.kind, NonterminalKind.AndExpression);
    }
    /**
     * Returns the child node that has the label `left_operand`.
     */
    get leftOperand() {
        return this.fetch().leftOperand;
    }
    /**
     * Returns the child node that has the label `operator`.
     */
    get operator() {
        return this.fetch().operator;
    }
    /**
     * Returns the child node that has the label `right_operand`.
     */
    get rightOperand() {
        return this.fetch().rightOperand;
    }
}
/**
 * This node represents a `EqualityExpression` nonterminal, with the following structure:
 *
 * ```ebnf
 * (* Left-associative binary operator *)
 * EqualityExpression = (* left_operand: *) Expression
 *                      (* operator: *) EQUAL_EQUAL
 *                      (* right_operand: *) Expression;
 *
 * (* Left-associative binary operator *)
 * EqualityExpression = (* left_operand: *) Expression
 *                      (* operator: *) BANG_EQUAL
 *                      (* right_operand: *) Expression;
 * ```
 */
export class EqualityExpression {
    /**
     * Constructs a new AST node of type `EqualityExpression`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `EqualityExpression`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const [$leftOperand, $operator, $rightOperand] = wasm.ast.Selectors.sequence(this.cst);
            return {
                leftOperand: new Expression($leftOperand),
                operator: $operator,
                rightOperand: new Expression($rightOperand),
            };
        });
        assertKind(this.cst.kind, NonterminalKind.EqualityExpression);
    }
    /**
     * Returns the child node that has the label `left_operand`.
     */
    get leftOperand() {
        return this.fetch().leftOperand;
    }
    /**
     * Returns the child node that has the label `operator`.
     */
    get operator() {
        return this.fetch().operator;
    }
    /**
     * Returns the child node that has the label `right_operand`.
     */
    get rightOperand() {
        return this.fetch().rightOperand;
    }
}
/**
 * This node represents a `InequalityExpression` nonterminal, with the following structure:
 *
 * ```ebnf
 * (* Left-associative binary operator *)
 * InequalityExpression = (* left_operand: *) Expression
 *                        (* operator: *) LESS_THAN
 *                        (* right_operand: *) Expression;
 *
 * (* Left-associative binary operator *)
 * InequalityExpression = (* left_operand: *) Expression
 *                        (* operator: *) GREATER_THAN
 *                        (* right_operand: *) Expression;
 *
 * (* Left-associative binary operator *)
 * InequalityExpression = (* left_operand: *) Expression
 *                        (* operator: *) LESS_THAN_EQUAL
 *                        (* right_operand: *) Expression;
 *
 * (* Left-associative binary operator *)
 * InequalityExpression = (* left_operand: *) Expression
 *                        (* operator: *) GREATER_THAN_EQUAL
 *                        (* right_operand: *) Expression;
 * ```
 */
export class InequalityExpression {
    /**
     * Constructs a new AST node of type `InequalityExpression`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `InequalityExpression`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const [$leftOperand, $operator, $rightOperand] = wasm.ast.Selectors.sequence(this.cst);
            return {
                leftOperand: new Expression($leftOperand),
                operator: $operator,
                rightOperand: new Expression($rightOperand),
            };
        });
        assertKind(this.cst.kind, NonterminalKind.InequalityExpression);
    }
    /**
     * Returns the child node that has the label `left_operand`.
     */
    get leftOperand() {
        return this.fetch().leftOperand;
    }
    /**
     * Returns the child node that has the label `operator`.
     */
    get operator() {
        return this.fetch().operator;
    }
    /**
     * Returns the child node that has the label `right_operand`.
     */
    get rightOperand() {
        return this.fetch().rightOperand;
    }
}
/**
 * This node represents a `BitwiseOrExpression` nonterminal, with the following structure:
 *
 * ```ebnf
 * (* Left-associative binary operator *)
 * BitwiseOrExpression = (* left_operand: *) Expression
 *                       (* operator: *) BAR
 *                       (* right_operand: *) Expression;
 * ```
 */
export class BitwiseOrExpression {
    /**
     * Constructs a new AST node of type `BitwiseOrExpression`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `BitwiseOrExpression`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const [$leftOperand, $operator, $rightOperand] = wasm.ast.Selectors.sequence(this.cst);
            return {
                leftOperand: new Expression($leftOperand),
                operator: $operator,
                rightOperand: new Expression($rightOperand),
            };
        });
        assertKind(this.cst.kind, NonterminalKind.BitwiseOrExpression);
    }
    /**
     * Returns the child node that has the label `left_operand`.
     */
    get leftOperand() {
        return this.fetch().leftOperand;
    }
    /**
     * Returns the child node that has the label `operator`.
     */
    get operator() {
        return this.fetch().operator;
    }
    /**
     * Returns the child node that has the label `right_operand`.
     */
    get rightOperand() {
        return this.fetch().rightOperand;
    }
}
/**
 * This node represents a `BitwiseXorExpression` nonterminal, with the following structure:
 *
 * ```ebnf
 * (* Left-associative binary operator *)
 * BitwiseXorExpression = (* left_operand: *) Expression
 *                        (* operator: *) CARET
 *                        (* right_operand: *) Expression;
 * ```
 */
export class BitwiseXorExpression {
    /**
     * Constructs a new AST node of type `BitwiseXorExpression`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `BitwiseXorExpression`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const [$leftOperand, $operator, $rightOperand] = wasm.ast.Selectors.sequence(this.cst);
            return {
                leftOperand: new Expression($leftOperand),
                operator: $operator,
                rightOperand: new Expression($rightOperand),
            };
        });
        assertKind(this.cst.kind, NonterminalKind.BitwiseXorExpression);
    }
    /**
     * Returns the child node that has the label `left_operand`.
     */
    get leftOperand() {
        return this.fetch().leftOperand;
    }
    /**
     * Returns the child node that has the label `operator`.
     */
    get operator() {
        return this.fetch().operator;
    }
    /**
     * Returns the child node that has the label `right_operand`.
     */
    get rightOperand() {
        return this.fetch().rightOperand;
    }
}
/**
 * This node represents a `BitwiseAndExpression` nonterminal, with the following structure:
 *
 * ```ebnf
 * (* Left-associative binary operator *)
 * BitwiseAndExpression = (* left_operand: *) Expression
 *                        (* operator: *) AMPERSAND
 *                        (* right_operand: *) Expression;
 * ```
 */
export class BitwiseAndExpression {
    /**
     * Constructs a new AST node of type `BitwiseAndExpression`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `BitwiseAndExpression`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const [$leftOperand, $operator, $rightOperand] = wasm.ast.Selectors.sequence(this.cst);
            return {
                leftOperand: new Expression($leftOperand),
                operator: $operator,
                rightOperand: new Expression($rightOperand),
            };
        });
        assertKind(this.cst.kind, NonterminalKind.BitwiseAndExpression);
    }
    /**
     * Returns the child node that has the label `left_operand`.
     */
    get leftOperand() {
        return this.fetch().leftOperand;
    }
    /**
     * Returns the child node that has the label `operator`.
     */
    get operator() {
        return this.fetch().operator;
    }
    /**
     * Returns the child node that has the label `right_operand`.
     */
    get rightOperand() {
        return this.fetch().rightOperand;
    }
}
/**
 * This node represents a `ShiftExpression` nonterminal, with the following structure:
 *
 * ```ebnf
 * (* Left-associative binary operator *)
 * ShiftExpression = (* left_operand: *) Expression
 *                   (* operator: *) LESS_THAN_LESS_THAN
 *                   (* right_operand: *) Expression;
 *
 * (* Left-associative binary operator *)
 * ShiftExpression = (* left_operand: *) Expression
 *                   (* operator: *) GREATER_THAN_GREATER_THAN
 *                   (* right_operand: *) Expression;
 *
 * (* Left-associative binary operator *)
 * ShiftExpression = (* left_operand: *) Expression
 *                   (* operator: *) GREATER_THAN_GREATER_THAN_GREATER_THAN
 *                   (* right_operand: *) Expression;
 * ```
 */
export class ShiftExpression {
    /**
     * Constructs a new AST node of type `ShiftExpression`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `ShiftExpression`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const [$leftOperand, $operator, $rightOperand] = wasm.ast.Selectors.sequence(this.cst);
            return {
                leftOperand: new Expression($leftOperand),
                operator: $operator,
                rightOperand: new Expression($rightOperand),
            };
        });
        assertKind(this.cst.kind, NonterminalKind.ShiftExpression);
    }
    /**
     * Returns the child node that has the label `left_operand`.
     */
    get leftOperand() {
        return this.fetch().leftOperand;
    }
    /**
     * Returns the child node that has the label `operator`.
     */
    get operator() {
        return this.fetch().operator;
    }
    /**
     * Returns the child node that has the label `right_operand`.
     */
    get rightOperand() {
        return this.fetch().rightOperand;
    }
}
/**
 * This node represents a `AdditiveExpression` nonterminal, with the following structure:
 *
 * ```ebnf
 * (* Left-associative binary operator *)
 * AdditiveExpression = (* left_operand: *) Expression
 *                      (* operator: *) PLUS
 *                      (* right_operand: *) Expression;
 *
 * (* Left-associative binary operator *)
 * AdditiveExpression = (* left_operand: *) Expression
 *                      (* operator: *) MINUS
 *                      (* right_operand: *) Expression;
 * ```
 */
export class AdditiveExpression {
    /**
     * Constructs a new AST node of type `AdditiveExpression`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `AdditiveExpression`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const [$leftOperand, $operator, $rightOperand] = wasm.ast.Selectors.sequence(this.cst);
            return {
                leftOperand: new Expression($leftOperand),
                operator: $operator,
                rightOperand: new Expression($rightOperand),
            };
        });
        assertKind(this.cst.kind, NonterminalKind.AdditiveExpression);
    }
    /**
     * Returns the child node that has the label `left_operand`.
     */
    get leftOperand() {
        return this.fetch().leftOperand;
    }
    /**
     * Returns the child node that has the label `operator`.
     */
    get operator() {
        return this.fetch().operator;
    }
    /**
     * Returns the child node that has the label `right_operand`.
     */
    get rightOperand() {
        return this.fetch().rightOperand;
    }
}
/**
 * This node represents a `MultiplicativeExpression` nonterminal, with the following structure:
 *
 * ```ebnf
 * (* Left-associative binary operator *)
 * MultiplicativeExpression = (* left_operand: *) Expression
 *                            (* operator: *) ASTERISK
 *                            (* right_operand: *) Expression;
 *
 * (* Left-associative binary operator *)
 * MultiplicativeExpression = (* left_operand: *) Expression
 *                            (* operator: *) SLASH
 *                            (* right_operand: *) Expression;
 *
 * (* Left-associative binary operator *)
 * MultiplicativeExpression = (* left_operand: *) Expression
 *                            (* operator: *) PERCENT
 *                            (* right_operand: *) Expression;
 * ```
 */
export class MultiplicativeExpression {
    /**
     * Constructs a new AST node of type `MultiplicativeExpression`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `MultiplicativeExpression`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const [$leftOperand, $operator, $rightOperand] = wasm.ast.Selectors.sequence(this.cst);
            return {
                leftOperand: new Expression($leftOperand),
                operator: $operator,
                rightOperand: new Expression($rightOperand),
            };
        });
        assertKind(this.cst.kind, NonterminalKind.MultiplicativeExpression);
    }
    /**
     * Returns the child node that has the label `left_operand`.
     */
    get leftOperand() {
        return this.fetch().leftOperand;
    }
    /**
     * Returns the child node that has the label `operator`.
     */
    get operator() {
        return this.fetch().operator;
    }
    /**
     * Returns the child node that has the label `right_operand`.
     */
    get rightOperand() {
        return this.fetch().rightOperand;
    }
}
/**
 * This node represents a `ExponentiationExpression` nonterminal, with the following structure:
 *
 * ```ebnf
 * (* Left-associative binary operator *)
 * (* Deprecated in 0.8.0 *)
 * ExponentiationExpression = (* left_operand: *) Expression
 *                            (* operator: *) ASTERISK_ASTERISK
 *                            (* right_operand: *) Expression;
 *
 * (* Right-associative binary operator *)
 * (* Introduced in 0.8.0 *)
 * ExponentiationExpression = (* left_operand: *) Expression
 *                            (* operator: *) ASTERISK_ASTERISK
 *                            (* right_operand: *) Expression;
 * ```
 */
export class ExponentiationExpression {
    /**
     * Constructs a new AST node of type `ExponentiationExpression`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `ExponentiationExpression`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const [$leftOperand, $operator, $rightOperand] = wasm.ast.Selectors.sequence(this.cst);
            return {
                leftOperand: new Expression($leftOperand),
                operator: $operator,
                rightOperand: new Expression($rightOperand),
            };
        });
        assertKind(this.cst.kind, NonterminalKind.ExponentiationExpression);
    }
    /**
     * Returns the child node that has the label `left_operand`.
     */
    get leftOperand() {
        return this.fetch().leftOperand;
    }
    /**
     * Returns the child node that has the label `operator`.
     */
    get operator() {
        return this.fetch().operator;
    }
    /**
     * Returns the child node that has the label `right_operand`.
     */
    get rightOperand() {
        return this.fetch().rightOperand;
    }
}
/**
 * This node represents a `PostfixExpression` nonterminal, with the following structure:
 *
 * ```ebnf
 * (* Postfix unary operator *)
 * PostfixExpression = (* operand: *) Expression
 *                     (* operator: *) PLUS_PLUS;
 *
 * (* Postfix unary operator *)
 * PostfixExpression = (* operand: *) Expression
 *                     (* operator: *) MINUS_MINUS;
 * ```
 */
export class PostfixExpression {
    /**
     * Constructs a new AST node of type `PostfixExpression`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `PostfixExpression`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const [$operand, $operator] = wasm.ast.Selectors.sequence(this.cst);
            return {
                operand: new Expression($operand),
                operator: $operator,
            };
        });
        assertKind(this.cst.kind, NonterminalKind.PostfixExpression);
    }
    /**
     * Returns the child node that has the label `operand`.
     */
    get operand() {
        return this.fetch().operand;
    }
    /**
     * Returns the child node that has the label `operator`.
     */
    get operator() {
        return this.fetch().operator;
    }
}
/**
 * This node represents a `PrefixExpression` nonterminal, with the following structure:
 *
 * ```ebnf
 * (* Prefix unary operator *)
 * PrefixExpression = (* operator: *) PLUS_PLUS
 *                    (* operand: *) Expression;
 *
 * (* Prefix unary operator *)
 * PrefixExpression = (* operator: *) MINUS_MINUS
 *                    (* operand: *) Expression;
 *
 * (* Prefix unary operator *)
 * PrefixExpression = (* operator: *) TILDE
 *                    (* operand: *) Expression;
 *
 * (* Prefix unary operator *)
 * PrefixExpression = (* operator: *) BANG
 *                    (* operand: *) Expression;
 *
 * (* Prefix unary operator *)
 * PrefixExpression = (* operator: *) MINUS
 *                    (* operand: *) Expression;
 *
 * (* Prefix unary operator *)
 * (* Deprecated in 0.5.0 *)
 * PrefixExpression = (* operator: *) PLUS
 *                    (* operand: *) Expression;
 *
 * (* Prefix unary operator *)
 * PrefixExpression = (* operator: *) DELETE_KEYWORD
 *                    (* operand: *) Expression;
 * ```
 */
export class PrefixExpression {
    /**
     * Constructs a new AST node of type `PrefixExpression`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `PrefixExpression`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const [$operator, $operand] = wasm.ast.Selectors.sequence(this.cst);
            return {
                operator: $operator,
                operand: new Expression($operand),
            };
        });
        assertKind(this.cst.kind, NonterminalKind.PrefixExpression);
    }
    /**
     * Returns the child node that has the label `operator`.
     */
    get operator() {
        return this.fetch().operator;
    }
    /**
     * Returns the child node that has the label `operand`.
     */
    get operand() {
        return this.fetch().operand;
    }
}
/**
 * This node represents a `FunctionCallExpression` nonterminal, with the following structure:
 *
 * ```ebnf
 * (* Postfix unary operator *)
 * FunctionCallExpression = (* operand: *) Expression
 *                          (* arguments: *) ArgumentsDeclaration;
 * ```
 */
export class FunctionCallExpression {
    /**
     * Constructs a new AST node of type `FunctionCallExpression`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `FunctionCallExpression`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const [$operand, $arguments] = wasm.ast.Selectors.sequence(this.cst);
            return {
                operand: new Expression($operand),
                arguments: new ArgumentsDeclaration($arguments),
            };
        });
        assertKind(this.cst.kind, NonterminalKind.FunctionCallExpression);
    }
    /**
     * Returns the child node that has the label `operand`.
     */
    get operand() {
        return this.fetch().operand;
    }
    /**
     * Returns the child node that has the label `arguments`.
     */
    get arguments() {
        return this.fetch().arguments;
    }
}
/**
 * This node represents a `CallOptionsExpression` nonterminal, with the following structure:
 *
 * ```ebnf
 * (* Postfix unary operator *)
 * (* Introduced in 0.6.2 *)
 * CallOptionsExpression = (* operand: *) Expression
 *                         (* open_brace: *) OPEN_BRACE
 *                         (* options: *) CallOptions
 *                         (* close_brace: *) CLOSE_BRACE;
 * ```
 */
export class CallOptionsExpression {
    /**
     * Constructs a new AST node of type `CallOptionsExpression`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `CallOptionsExpression`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const [$operand, $openBrace, $options, $closeBrace] = wasm.ast.Selectors.sequence(this.cst);
            return {
                operand: new Expression($operand),
                openBrace: $openBrace,
                options: new CallOptions($options),
                closeBrace: $closeBrace,
            };
        });
        assertKind(this.cst.kind, NonterminalKind.CallOptionsExpression);
    }
    /**
     * Returns the child node that has the label `operand`.
     */
    get operand() {
        return this.fetch().operand;
    }
    /**
     * Returns the child node that has the label `open_brace`.
     */
    get openBrace() {
        return this.fetch().openBrace;
    }
    /**
     * Returns the child node that has the label `options`.
     */
    get options() {
        return this.fetch().options;
    }
    /**
     * Returns the child node that has the label `close_brace`.
     */
    get closeBrace() {
        return this.fetch().closeBrace;
    }
}
/**
 * This node represents a `MemberAccessExpression` nonterminal, with the following structure:
 *
 * ```ebnf
 * (* Postfix unary operator *)
 * MemberAccessExpression = (* operand: *) Expression
 *                          (* period: *) PERIOD
 *                          (* member: *) IDENTIFIER;
 * ```
 */
export class MemberAccessExpression {
    /**
     * Constructs a new AST node of type `MemberAccessExpression`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `MemberAccessExpression`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const [$operand, $period, $member] = wasm.ast.Selectors.sequence(this.cst);
            return {
                operand: new Expression($operand),
                period: $period,
                member: $member,
            };
        });
        assertKind(this.cst.kind, NonterminalKind.MemberAccessExpression);
    }
    /**
     * Returns the child node that has the label `operand`.
     */
    get operand() {
        return this.fetch().operand;
    }
    /**
     * Returns the child node that has the label `period`.
     */
    get period() {
        return this.fetch().period;
    }
    /**
     * Returns the child node that has the label `member`.
     */
    get member() {
        return this.fetch().member;
    }
}
/**
 * This node represents a `IndexAccessExpression` nonterminal, with the following structure:
 *
 * ```ebnf
 * (* Postfix unary operator *)
 * IndexAccessExpression = (* operand: *) Expression
 *                         (* open_bracket: *) OPEN_BRACKET
 *                         (* start: *) Expression?
 *                         (* end: *) IndexAccessEnd?
 *                         (* close_bracket: *) CLOSE_BRACKET;
 * ```
 */
export class IndexAccessExpression {
    /**
     * Constructs a new AST node of type `IndexAccessExpression`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `IndexAccessExpression`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const [$operand, $openBracket, $start, $end, $closeBracket] = wasm.ast.Selectors.sequence(this.cst);
            return {
                operand: new Expression($operand),
                openBracket: $openBracket,
                start: $start === undefined ? undefined : new Expression($start),
                end: $end === undefined ? undefined : new IndexAccessEnd($end),
                closeBracket: $closeBracket,
            };
        });
        assertKind(this.cst.kind, NonterminalKind.IndexAccessExpression);
    }
    /**
     * Returns the child node that has the label `operand`.
     */
    get operand() {
        return this.fetch().operand;
    }
    /**
     * Returns the child node that has the label `open_bracket`.
     */
    get openBracket() {
        return this.fetch().openBracket;
    }
    /**
     * Returns the child node that has the label `start`.
     */
    get start() {
        return this.fetch().start;
    }
    /**
     * Returns the child node that has the label `end`.
     */
    get end() {
        return this.fetch().end;
    }
    /**
     * Returns the child node that has the label `close_bracket`.
     */
    get closeBracket() {
        return this.fetch().closeBracket;
    }
}
/**
 * This node represents a `IndexAccessEnd` nonterminal, with the following structure:
 *
 * ```ebnf
 * IndexAccessEnd = (* colon: *) COLON
 *                  (* end: *) Expression?;
 * ```
 */
export class IndexAccessEnd {
    /**
     * Constructs a new AST node of type `IndexAccessEnd`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `IndexAccessEnd`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const [$colon, $end] = wasm.ast.Selectors.sequence(this.cst);
            return {
                colon: $colon,
                end: $end === undefined ? undefined : new Expression($end),
            };
        });
        assertKind(this.cst.kind, NonterminalKind.IndexAccessEnd);
    }
    /**
     * Returns the child node that has the label `colon`.
     */
    get colon() {
        return this.fetch().colon;
    }
    /**
     * Returns the child node that has the label `end`.
     */
    get end() {
        return this.fetch().end;
    }
}
/**
 * This node represents a `PositionalArgumentsDeclaration` nonterminal, with the following structure:
 *
 * ```ebnf
 * PositionalArgumentsDeclaration = (* open_paren: *) OPEN_PAREN
 *                                  (* arguments: *) PositionalArguments
 *                                  (* close_paren: *) CLOSE_PAREN;
 * ```
 */
export class PositionalArgumentsDeclaration {
    /**
     * Constructs a new AST node of type `PositionalArgumentsDeclaration`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `PositionalArgumentsDeclaration`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const [$openParen, $arguments, $closeParen] = wasm.ast.Selectors.sequence(this.cst);
            return {
                openParen: $openParen,
                arguments: new PositionalArguments($arguments),
                closeParen: $closeParen,
            };
        });
        assertKind(this.cst.kind, NonterminalKind.PositionalArgumentsDeclaration);
    }
    /**
     * Returns the child node that has the label `open_paren`.
     */
    get openParen() {
        return this.fetch().openParen;
    }
    /**
     * Returns the child node that has the label `arguments`.
     */
    get arguments() {
        return this.fetch().arguments;
    }
    /**
     * Returns the child node that has the label `close_paren`.
     */
    get closeParen() {
        return this.fetch().closeParen;
    }
}
/**
 * This node represents a `NamedArgumentsDeclaration` nonterminal, with the following structure:
 *
 * ```ebnf
 * NamedArgumentsDeclaration = (* open_paren: *) OPEN_PAREN
 *                             (* arguments: *) NamedArgumentGroup?
 *                             (* close_paren: *) CLOSE_PAREN;
 * ```
 */
export class NamedArgumentsDeclaration {
    /**
     * Constructs a new AST node of type `NamedArgumentsDeclaration`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `NamedArgumentsDeclaration`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const [$openParen, $arguments, $closeParen] = wasm.ast.Selectors.sequence(this.cst);
            return {
                openParen: $openParen,
                arguments: $arguments === undefined ? undefined : new NamedArgumentGroup($arguments),
                closeParen: $closeParen,
            };
        });
        assertKind(this.cst.kind, NonterminalKind.NamedArgumentsDeclaration);
    }
    /**
     * Returns the child node that has the label `open_paren`.
     */
    get openParen() {
        return this.fetch().openParen;
    }
    /**
     * Returns the child node that has the label `arguments`.
     */
    get arguments() {
        return this.fetch().arguments;
    }
    /**
     * Returns the child node that has the label `close_paren`.
     */
    get closeParen() {
        return this.fetch().closeParen;
    }
}
/**
 * This node represents a `NamedArgumentGroup` nonterminal, with the following structure:
 *
 * ```ebnf
 * NamedArgumentGroup = (* open_brace: *) OPEN_BRACE
 *                      (* arguments: *) NamedArguments
 *                      (* close_brace: *) CLOSE_BRACE;
 * ```
 */
export class NamedArgumentGroup {
    /**
     * Constructs a new AST node of type `NamedArgumentGroup`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `NamedArgumentGroup`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const [$openBrace, $arguments, $closeBrace] = wasm.ast.Selectors.sequence(this.cst);
            return {
                openBrace: $openBrace,
                arguments: new NamedArguments($arguments),
                closeBrace: $closeBrace,
            };
        });
        assertKind(this.cst.kind, NonterminalKind.NamedArgumentGroup);
    }
    /**
     * Returns the child node that has the label `open_brace`.
     */
    get openBrace() {
        return this.fetch().openBrace;
    }
    /**
     * Returns the child node that has the label `arguments`.
     */
    get arguments() {
        return this.fetch().arguments;
    }
    /**
     * Returns the child node that has the label `close_brace`.
     */
    get closeBrace() {
        return this.fetch().closeBrace;
    }
}
/**
 * This node represents a `NamedArgument` nonterminal, with the following structure:
 *
 * ```ebnf
 * NamedArgument = (* name: *) IDENTIFIER
 *                 (* colon: *) COLON
 *                 (* value: *) Expression;
 * ```
 */
export class NamedArgument {
    /**
     * Constructs a new AST node of type `NamedArgument`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `NamedArgument`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const [$name, $colon, $value] = wasm.ast.Selectors.sequence(this.cst);
            return {
                name: $name,
                colon: $colon,
                value: new Expression($value),
            };
        });
        assertKind(this.cst.kind, NonterminalKind.NamedArgument);
    }
    /**
     * Returns the child node that has the label `name`.
     */
    get name() {
        return this.fetch().name;
    }
    /**
     * Returns the child node that has the label `colon`.
     */
    get colon() {
        return this.fetch().colon;
    }
    /**
     * Returns the child node that has the label `value`.
     */
    get value() {
        return this.fetch().value;
    }
}
/**
 * This node represents a `TypeExpression` nonterminal, with the following structure:
 *
 * ```ebnf
 * (* Introduced in 0.5.3 *)
 * TypeExpression = (* type_keyword: *) TYPE_KEYWORD
 *                  (* open_paren: *) OPEN_PAREN
 *                  (* type_name: *) TypeName
 *                  (* close_paren: *) CLOSE_PAREN;
 * ```
 */
export class TypeExpression {
    /**
     * Constructs a new AST node of type `TypeExpression`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `TypeExpression`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const [$typeKeyword, $openParen, $typeName, $closeParen] = wasm.ast.Selectors.sequence(this.cst);
            return {
                typeKeyword: $typeKeyword,
                openParen: $openParen,
                typeName: new TypeName($typeName),
                closeParen: $closeParen,
            };
        });
        assertKind(this.cst.kind, NonterminalKind.TypeExpression);
    }
    /**
     * Returns the child node that has the label `type_keyword`.
     */
    get typeKeyword() {
        return this.fetch().typeKeyword;
    }
    /**
     * Returns the child node that has the label `open_paren`.
     */
    get openParen() {
        return this.fetch().openParen;
    }
    /**
     * Returns the child node that has the label `type_name`.
     */
    get typeName() {
        return this.fetch().typeName;
    }
    /**
     * Returns the child node that has the label `close_paren`.
     */
    get closeParen() {
        return this.fetch().closeParen;
    }
}
/**
 * This node represents a `NewExpression` nonterminal, with the following structure:
 *
 * ```ebnf
 * NewExpression = (* new_keyword: *) NEW_KEYWORD
 *                 (* type_name: *) TypeName;
 * ```
 */
export class NewExpression {
    /**
     * Constructs a new AST node of type `NewExpression`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `NewExpression`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const [$newKeyword, $typeName] = wasm.ast.Selectors.sequence(this.cst);
            return {
                newKeyword: $newKeyword,
                typeName: new TypeName($typeName),
            };
        });
        assertKind(this.cst.kind, NonterminalKind.NewExpression);
    }
    /**
     * Returns the child node that has the label `new_keyword`.
     */
    get newKeyword() {
        return this.fetch().newKeyword;
    }
    /**
     * Returns the child node that has the label `type_name`.
     */
    get typeName() {
        return this.fetch().typeName;
    }
}
/**
 * This node represents a `TupleExpression` nonterminal, with the following structure:
 *
 * ```ebnf
 * TupleExpression = (* open_paren: *) OPEN_PAREN
 *                   (* items: *) TupleValues
 *                   (* close_paren: *) CLOSE_PAREN;
 * ```
 */
export class TupleExpression {
    /**
     * Constructs a new AST node of type `TupleExpression`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `TupleExpression`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const [$openParen, $items, $closeParen] = wasm.ast.Selectors.sequence(this.cst);
            return {
                openParen: $openParen,
                items: new TupleValues($items),
                closeParen: $closeParen,
            };
        });
        assertKind(this.cst.kind, NonterminalKind.TupleExpression);
    }
    /**
     * Returns the child node that has the label `open_paren`.
     */
    get openParen() {
        return this.fetch().openParen;
    }
    /**
     * Returns the child node that has the label `items`.
     */
    get items() {
        return this.fetch().items;
    }
    /**
     * Returns the child node that has the label `close_paren`.
     */
    get closeParen() {
        return this.fetch().closeParen;
    }
}
/**
 * This node represents a `TupleValue` nonterminal, with the following structure:
 *
 * ```ebnf
 * TupleValue = (* expression: *) Expression?;
 * ```
 */
export class TupleValue {
    /**
     * Constructs a new AST node of type `TupleValue`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `TupleValue`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const [$expression] = wasm.ast.Selectors.sequence(this.cst);
            return {
                expression: $expression === undefined ? undefined : new Expression($expression),
            };
        });
        assertKind(this.cst.kind, NonterminalKind.TupleValue);
    }
    /**
     * Returns the child node that has the label `expression`.
     */
    get expression() {
        return this.fetch().expression;
    }
}
/**
 * This node represents a `ArrayExpression` nonterminal, with the following structure:
 *
 * ```ebnf
 * ArrayExpression = (* open_bracket: *) OPEN_BRACKET
 *                   (* items: *) ArrayValues
 *                   (* close_bracket: *) CLOSE_BRACKET;
 * ```
 */
export class ArrayExpression {
    /**
     * Constructs a new AST node of type `ArrayExpression`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `ArrayExpression`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const [$openBracket, $items, $closeBracket] = wasm.ast.Selectors.sequence(this.cst);
            return {
                openBracket: $openBracket,
                items: new ArrayValues($items),
                closeBracket: $closeBracket,
            };
        });
        assertKind(this.cst.kind, NonterminalKind.ArrayExpression);
    }
    /**
     * Returns the child node that has the label `open_bracket`.
     */
    get openBracket() {
        return this.fetch().openBracket;
    }
    /**
     * Returns the child node that has the label `items`.
     */
    get items() {
        return this.fetch().items;
    }
    /**
     * Returns the child node that has the label `close_bracket`.
     */
    get closeBracket() {
        return this.fetch().closeBracket;
    }
}
/**
 * This node represents a `HexNumberExpression` nonterminal, with the following structure:
 *
 * ```ebnf
 * HexNumberExpression = (* literal: *) HEX_LITERAL
 *                       (* unit: *) NumberUnit?; (* Deprecated in 0.5.0 *)
 * ```
 */
export class HexNumberExpression {
    /**
     * Constructs a new AST node of type `HexNumberExpression`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `HexNumberExpression`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const [$literal, $unit] = wasm.ast.Selectors.sequence(this.cst);
            return {
                literal: $literal,
                unit: $unit === undefined ? undefined : new NumberUnit($unit),
            };
        });
        assertKind(this.cst.kind, NonterminalKind.HexNumberExpression);
    }
    /**
     * Returns the child node that has the label `literal`.
     */
    get literal() {
        return this.fetch().literal;
    }
    /**
     * Returns the child node that has the label `unit`.
     */
    get unit() {
        return this.fetch().unit;
    }
}
/**
 * This node represents a `DecimalNumberExpression` nonterminal, with the following structure:
 *
 * ```ebnf
 * DecimalNumberExpression = (* literal: *) DECIMAL_LITERAL
 *                           (* unit: *) NumberUnit?;
 * ```
 */
export class DecimalNumberExpression {
    /**
     * Constructs a new AST node of type `DecimalNumberExpression`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `DecimalNumberExpression`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const [$literal, $unit] = wasm.ast.Selectors.sequence(this.cst);
            return {
                literal: $literal,
                unit: $unit === undefined ? undefined : new NumberUnit($unit),
            };
        });
        assertKind(this.cst.kind, NonterminalKind.DecimalNumberExpression);
    }
    /**
     * Returns the child node that has the label `literal`.
     */
    get literal() {
        return this.fetch().literal;
    }
    /**
     * Returns the child node that has the label `unit`.
     */
    get unit() {
        return this.fetch().unit;
    }
}
/**
 * This node represents a `YulBlock` nonterminal, with the following structure:
 *
 * ```ebnf
 * YulBlock = (* open_brace: *) OPEN_BRACE
 *            (* statements: *) YulStatements
 *            (* close_brace: *) CLOSE_BRACE;
 * ```
 */
export class YulBlock {
    /**
     * Constructs a new AST node of type `YulBlock`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `YulBlock`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const [$openBrace, $statements, $closeBrace] = wasm.ast.Selectors.sequence(this.cst);
            return {
                openBrace: $openBrace,
                statements: new YulStatements($statements),
                closeBrace: $closeBrace,
            };
        });
        assertKind(this.cst.kind, NonterminalKind.YulBlock);
    }
    /**
     * Returns the child node that has the label `open_brace`.
     */
    get openBrace() {
        return this.fetch().openBrace;
    }
    /**
     * Returns the child node that has the label `statements`.
     */
    get statements() {
        return this.fetch().statements;
    }
    /**
     * Returns the child node that has the label `close_brace`.
     */
    get closeBrace() {
        return this.fetch().closeBrace;
    }
}
/**
 * This node represents a `YulFunctionDefinition` nonterminal, with the following structure:
 *
 * ```ebnf
 * YulFunctionDefinition = (* function_keyword: *) YUL_FUNCTION_KEYWORD
 *                         (* name: *) YUL_IDENTIFIER
 *                         (* parameters: *) YulParametersDeclaration
 *                         (* returns: *) YulReturnsDeclaration?
 *                         (* body: *) YulBlock;
 * ```
 */
export class YulFunctionDefinition {
    /**
     * Constructs a new AST node of type `YulFunctionDefinition`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `YulFunctionDefinition`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const [$functionKeyword, $name, $parameters, $returns, $body] = wasm.ast.Selectors.sequence(this.cst);
            return {
                functionKeyword: $functionKeyword,
                name: $name,
                parameters: new YulParametersDeclaration($parameters),
                returns: $returns === undefined ? undefined : new YulReturnsDeclaration($returns),
                body: new YulBlock($body),
            };
        });
        assertKind(this.cst.kind, NonterminalKind.YulFunctionDefinition);
    }
    /**
     * Returns the child node that has the label `function_keyword`.
     */
    get functionKeyword() {
        return this.fetch().functionKeyword;
    }
    /**
     * Returns the child node that has the label `name`.
     */
    get name() {
        return this.fetch().name;
    }
    /**
     * Returns the child node that has the label `parameters`.
     */
    get parameters() {
        return this.fetch().parameters;
    }
    /**
     * Returns the child node that has the label `returns`.
     */
    get returns() {
        return this.fetch().returns;
    }
    /**
     * Returns the child node that has the label `body`.
     */
    get body() {
        return this.fetch().body;
    }
}
/**
 * This node represents a `YulParametersDeclaration` nonterminal, with the following structure:
 *
 * ```ebnf
 * YulParametersDeclaration = (* open_paren: *) OPEN_PAREN
 *                            (* parameters: *) YulParameters
 *                            (* close_paren: *) CLOSE_PAREN;
 * ```
 */
export class YulParametersDeclaration {
    /**
     * Constructs a new AST node of type `YulParametersDeclaration`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `YulParametersDeclaration`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const [$openParen, $parameters, $closeParen] = wasm.ast.Selectors.sequence(this.cst);
            return {
                openParen: $openParen,
                parameters: new YulParameters($parameters),
                closeParen: $closeParen,
            };
        });
        assertKind(this.cst.kind, NonterminalKind.YulParametersDeclaration);
    }
    /**
     * Returns the child node that has the label `open_paren`.
     */
    get openParen() {
        return this.fetch().openParen;
    }
    /**
     * Returns the child node that has the label `parameters`.
     */
    get parameters() {
        return this.fetch().parameters;
    }
    /**
     * Returns the child node that has the label `close_paren`.
     */
    get closeParen() {
        return this.fetch().closeParen;
    }
}
/**
 * This node represents a `YulReturnsDeclaration` nonterminal, with the following structure:
 *
 * ```ebnf
 * YulReturnsDeclaration = (* minus_greater_than: *) MINUS_GREATER_THAN
 *                         (* variables: *) YulVariableNames;
 * ```
 */
export class YulReturnsDeclaration {
    /**
     * Constructs a new AST node of type `YulReturnsDeclaration`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `YulReturnsDeclaration`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const [$minusGreaterThan, $variables] = wasm.ast.Selectors.sequence(this.cst);
            return {
                minusGreaterThan: $minusGreaterThan,
                variables: new YulVariableNames($variables),
            };
        });
        assertKind(this.cst.kind, NonterminalKind.YulReturnsDeclaration);
    }
    /**
     * Returns the child node that has the label `minus_greater_than`.
     */
    get minusGreaterThan() {
        return this.fetch().minusGreaterThan;
    }
    /**
     * Returns the child node that has the label `variables`.
     */
    get variables() {
        return this.fetch().variables;
    }
}
/**
 * This node represents a `YulVariableDeclarationStatement` nonterminal, with the following structure:
 *
 * ```ebnf
 * YulVariableDeclarationStatement = (* let_keyword: *) YUL_LET_KEYWORD
 *                                   (* variables: *) YulVariableNames
 *                                   (* value: *) YulVariableDeclarationValue?;
 * ```
 */
export class YulVariableDeclarationStatement {
    /**
     * Constructs a new AST node of type `YulVariableDeclarationStatement`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `YulVariableDeclarationStatement`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const [$letKeyword, $variables, $value] = wasm.ast.Selectors.sequence(this.cst);
            return {
                letKeyword: $letKeyword,
                variables: new YulVariableNames($variables),
                value: $value === undefined ? undefined : new YulVariableDeclarationValue($value),
            };
        });
        assertKind(this.cst.kind, NonterminalKind.YulVariableDeclarationStatement);
    }
    /**
     * Returns the child node that has the label `let_keyword`.
     */
    get letKeyword() {
        return this.fetch().letKeyword;
    }
    /**
     * Returns the child node that has the label `variables`.
     */
    get variables() {
        return this.fetch().variables;
    }
    /**
     * Returns the child node that has the label `value`.
     */
    get value() {
        return this.fetch().value;
    }
}
/**
 * This node represents a `YulVariableDeclarationValue` nonterminal, with the following structure:
 *
 * ```ebnf
 * YulVariableDeclarationValue = (* assignment: *) YulAssignmentOperator
 *                               (* expression: *) YulExpression;
 * ```
 */
export class YulVariableDeclarationValue {
    /**
     * Constructs a new AST node of type `YulVariableDeclarationValue`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `YulVariableDeclarationValue`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const [$assignment, $expression] = wasm.ast.Selectors.sequence(this.cst);
            return {
                assignment: new YulAssignmentOperator($assignment),
                expression: new YulExpression($expression),
            };
        });
        assertKind(this.cst.kind, NonterminalKind.YulVariableDeclarationValue);
    }
    /**
     * Returns the child node that has the label `assignment`.
     */
    get assignment() {
        return this.fetch().assignment;
    }
    /**
     * Returns the child node that has the label `expression`.
     */
    get expression() {
        return this.fetch().expression;
    }
}
/**
 * This node represents a `YulVariableAssignmentStatement` nonterminal, with the following structure:
 *
 * ```ebnf
 * YulVariableAssignmentStatement = (* variables: *) YulPaths
 *                                  (* assignment: *) YulAssignmentOperator
 *                                  (* expression: *) YulExpression;
 * ```
 */
export class YulVariableAssignmentStatement {
    /**
     * Constructs a new AST node of type `YulVariableAssignmentStatement`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `YulVariableAssignmentStatement`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const [$variables, $assignment, $expression] = wasm.ast.Selectors.sequence(this.cst);
            return {
                variables: new YulPaths($variables),
                assignment: new YulAssignmentOperator($assignment),
                expression: new YulExpression($expression),
            };
        });
        assertKind(this.cst.kind, NonterminalKind.YulVariableAssignmentStatement);
    }
    /**
     * Returns the child node that has the label `variables`.
     */
    get variables() {
        return this.fetch().variables;
    }
    /**
     * Returns the child node that has the label `assignment`.
     */
    get assignment() {
        return this.fetch().assignment;
    }
    /**
     * Returns the child node that has the label `expression`.
     */
    get expression() {
        return this.fetch().expression;
    }
}
/**
 * This node represents a `YulColonAndEqual` nonterminal, with the following structure:
 *
 * ```ebnf
 * (* Deprecated in 0.5.5 *)
 * YulColonAndEqual = (* colon: *) COLON
 *                    (* equal: *) EQUAL;
 * ```
 */
export class YulColonAndEqual {
    /**
     * Constructs a new AST node of type `YulColonAndEqual`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `YulColonAndEqual`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const [$colon, $equal] = wasm.ast.Selectors.sequence(this.cst);
            return {
                colon: $colon,
                equal: $equal,
            };
        });
        assertKind(this.cst.kind, NonterminalKind.YulColonAndEqual);
    }
    /**
     * Returns the child node that has the label `colon`.
     */
    get colon() {
        return this.fetch().colon;
    }
    /**
     * Returns the child node that has the label `equal`.
     */
    get equal() {
        return this.fetch().equal;
    }
}
/**
 * This node represents a `YulStackAssignmentStatement` nonterminal, with the following structure:
 *
 * ```ebnf
 * (* Deprecated in 0.5.0 *)
 * YulStackAssignmentStatement = (* assignment: *) YulStackAssignmentOperator
 *                               (* variable: *) YUL_IDENTIFIER;
 * ```
 */
export class YulStackAssignmentStatement {
    /**
     * Constructs a new AST node of type `YulStackAssignmentStatement`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `YulStackAssignmentStatement`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const [$assignment, $variable] = wasm.ast.Selectors.sequence(this.cst);
            return {
                assignment: new YulStackAssignmentOperator($assignment),
                variable: $variable,
            };
        });
        assertKind(this.cst.kind, NonterminalKind.YulStackAssignmentStatement);
    }
    /**
     * Returns the child node that has the label `assignment`.
     */
    get assignment() {
        return this.fetch().assignment;
    }
    /**
     * Returns the child node that has the label `variable`.
     */
    get variable() {
        return this.fetch().variable;
    }
}
/**
 * This node represents a `YulEqualAndColon` nonterminal, with the following structure:
 *
 * ```ebnf
 * (* Deprecated in 0.5.0 *)
 * YulEqualAndColon = (* equal: *) EQUAL
 *                    (* colon: *) COLON;
 * ```
 */
export class YulEqualAndColon {
    /**
     * Constructs a new AST node of type `YulEqualAndColon`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `YulEqualAndColon`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const [$equal, $colon] = wasm.ast.Selectors.sequence(this.cst);
            return {
                equal: $equal,
                colon: $colon,
            };
        });
        assertKind(this.cst.kind, NonterminalKind.YulEqualAndColon);
    }
    /**
     * Returns the child node that has the label `equal`.
     */
    get equal() {
        return this.fetch().equal;
    }
    /**
     * Returns the child node that has the label `colon`.
     */
    get colon() {
        return this.fetch().colon;
    }
}
/**
 * This node represents a `YulIfStatement` nonterminal, with the following structure:
 *
 * ```ebnf
 * YulIfStatement = (* if_keyword: *) YUL_IF_KEYWORD
 *                  (* condition: *) YulExpression
 *                  (* body: *) YulBlock;
 * ```
 */
export class YulIfStatement {
    /**
     * Constructs a new AST node of type `YulIfStatement`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `YulIfStatement`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const [$ifKeyword, $condition, $body] = wasm.ast.Selectors.sequence(this.cst);
            return {
                ifKeyword: $ifKeyword,
                condition: new YulExpression($condition),
                body: new YulBlock($body),
            };
        });
        assertKind(this.cst.kind, NonterminalKind.YulIfStatement);
    }
    /**
     * Returns the child node that has the label `if_keyword`.
     */
    get ifKeyword() {
        return this.fetch().ifKeyword;
    }
    /**
     * Returns the child node that has the label `condition`.
     */
    get condition() {
        return this.fetch().condition;
    }
    /**
     * Returns the child node that has the label `body`.
     */
    get body() {
        return this.fetch().body;
    }
}
/**
 * This node represents a `YulForStatement` nonterminal, with the following structure:
 *
 * ```ebnf
 * YulForStatement = (* for_keyword: *) YUL_FOR_KEYWORD
 *                   (* initialization: *) YulBlock
 *                   (* condition: *) YulExpression
 *                   (* iterator: *) YulBlock
 *                   (* body: *) YulBlock;
 * ```
 */
export class YulForStatement {
    /**
     * Constructs a new AST node of type `YulForStatement`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `YulForStatement`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const [$forKeyword, $initialization, $condition, $iterator, $body] = wasm.ast.Selectors.sequence(this.cst);
            return {
                forKeyword: $forKeyword,
                initialization: new YulBlock($initialization),
                condition: new YulExpression($condition),
                iterator: new YulBlock($iterator),
                body: new YulBlock($body),
            };
        });
        assertKind(this.cst.kind, NonterminalKind.YulForStatement);
    }
    /**
     * Returns the child node that has the label `for_keyword`.
     */
    get forKeyword() {
        return this.fetch().forKeyword;
    }
    /**
     * Returns the child node that has the label `initialization`.
     */
    get initialization() {
        return this.fetch().initialization;
    }
    /**
     * Returns the child node that has the label `condition`.
     */
    get condition() {
        return this.fetch().condition;
    }
    /**
     * Returns the child node that has the label `iterator`.
     */
    get iterator() {
        return this.fetch().iterator;
    }
    /**
     * Returns the child node that has the label `body`.
     */
    get body() {
        return this.fetch().body;
    }
}
/**
 * This node represents a `YulSwitchStatement` nonterminal, with the following structure:
 *
 * ```ebnf
 * YulSwitchStatement = (* switch_keyword: *) YUL_SWITCH_KEYWORD
 *                      (* expression: *) YulExpression
 *                      (* cases: *) YulSwitchCases;
 * ```
 */
export class YulSwitchStatement {
    /**
     * Constructs a new AST node of type `YulSwitchStatement`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `YulSwitchStatement`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const [$switchKeyword, $expression, $cases] = wasm.ast.Selectors.sequence(this.cst);
            return {
                switchKeyword: $switchKeyword,
                expression: new YulExpression($expression),
                cases: new YulSwitchCases($cases),
            };
        });
        assertKind(this.cst.kind, NonterminalKind.YulSwitchStatement);
    }
    /**
     * Returns the child node that has the label `switch_keyword`.
     */
    get switchKeyword() {
        return this.fetch().switchKeyword;
    }
    /**
     * Returns the child node that has the label `expression`.
     */
    get expression() {
        return this.fetch().expression;
    }
    /**
     * Returns the child node that has the label `cases`.
     */
    get cases() {
        return this.fetch().cases;
    }
}
/**
 * This node represents a `YulDefaultCase` nonterminal, with the following structure:
 *
 * ```ebnf
 * YulDefaultCase = (* default_keyword: *) YUL_DEFAULT_KEYWORD
 *                  (* body: *) YulBlock;
 * ```
 */
export class YulDefaultCase {
    /**
     * Constructs a new AST node of type `YulDefaultCase`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `YulDefaultCase`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const [$defaultKeyword, $body] = wasm.ast.Selectors.sequence(this.cst);
            return {
                defaultKeyword: $defaultKeyword,
                body: new YulBlock($body),
            };
        });
        assertKind(this.cst.kind, NonterminalKind.YulDefaultCase);
    }
    /**
     * Returns the child node that has the label `default_keyword`.
     */
    get defaultKeyword() {
        return this.fetch().defaultKeyword;
    }
    /**
     * Returns the child node that has the label `body`.
     */
    get body() {
        return this.fetch().body;
    }
}
/**
 * This node represents a `YulValueCase` nonterminal, with the following structure:
 *
 * ```ebnf
 * YulValueCase = (* case_keyword: *) YUL_CASE_KEYWORD
 *                (* value: *) YulLiteral
 *                (* body: *) YulBlock;
 * ```
 */
export class YulValueCase {
    /**
     * Constructs a new AST node of type `YulValueCase`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `YulValueCase`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const [$caseKeyword, $value, $body] = wasm.ast.Selectors.sequence(this.cst);
            return {
                caseKeyword: $caseKeyword,
                value: new YulLiteral($value),
                body: new YulBlock($body),
            };
        });
        assertKind(this.cst.kind, NonterminalKind.YulValueCase);
    }
    /**
     * Returns the child node that has the label `case_keyword`.
     */
    get caseKeyword() {
        return this.fetch().caseKeyword;
    }
    /**
     * Returns the child node that has the label `value`.
     */
    get value() {
        return this.fetch().value;
    }
    /**
     * Returns the child node that has the label `body`.
     */
    get body() {
        return this.fetch().body;
    }
}
/**
 * This node represents a `YulLeaveStatement` nonterminal, with the following structure:
 *
 * ```ebnf
 * (* Introduced in 0.6.0 *)
 * YulLeaveStatement = (* leave_keyword: *) YUL_LEAVE_KEYWORD;
 * ```
 */
export class YulLeaveStatement {
    /**
     * Constructs a new AST node of type `YulLeaveStatement`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `YulLeaveStatement`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const [$leaveKeyword] = wasm.ast.Selectors.sequence(this.cst);
            return {
                leaveKeyword: $leaveKeyword,
            };
        });
        assertKind(this.cst.kind, NonterminalKind.YulLeaveStatement);
    }
    /**
     * Returns the child node that has the label `leave_keyword`.
     */
    get leaveKeyword() {
        return this.fetch().leaveKeyword;
    }
}
/**
 * This node represents a `YulBreakStatement` nonterminal, with the following structure:
 *
 * ```ebnf
 * YulBreakStatement = (* break_keyword: *) YUL_BREAK_KEYWORD;
 * ```
 */
export class YulBreakStatement {
    /**
     * Constructs a new AST node of type `YulBreakStatement`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `YulBreakStatement`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const [$breakKeyword] = wasm.ast.Selectors.sequence(this.cst);
            return {
                breakKeyword: $breakKeyword,
            };
        });
        assertKind(this.cst.kind, NonterminalKind.YulBreakStatement);
    }
    /**
     * Returns the child node that has the label `break_keyword`.
     */
    get breakKeyword() {
        return this.fetch().breakKeyword;
    }
}
/**
 * This node represents a `YulContinueStatement` nonterminal, with the following structure:
 *
 * ```ebnf
 * YulContinueStatement = (* continue_keyword: *) YUL_CONTINUE_KEYWORD;
 * ```
 */
export class YulContinueStatement {
    /**
     * Constructs a new AST node of type `YulContinueStatement`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `YulContinueStatement`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const [$continueKeyword] = wasm.ast.Selectors.sequence(this.cst);
            return {
                continueKeyword: $continueKeyword,
            };
        });
        assertKind(this.cst.kind, NonterminalKind.YulContinueStatement);
    }
    /**
     * Returns the child node that has the label `continue_keyword`.
     */
    get continueKeyword() {
        return this.fetch().continueKeyword;
    }
}
/**
 * This node represents a `YulLabel` nonterminal, with the following structure:
 *
 * ```ebnf
 * (* Deprecated in 0.5.0 *)
 * YulLabel = (* label: *) YUL_IDENTIFIER
 *            (* colon: *) COLON;
 * ```
 */
export class YulLabel {
    /**
     * Constructs a new AST node of type `YulLabel`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `YulLabel`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const [$label, $colon] = wasm.ast.Selectors.sequence(this.cst);
            return {
                label: $label,
                colon: $colon,
            };
        });
        assertKind(this.cst.kind, NonterminalKind.YulLabel);
    }
    /**
     * Returns the child node that has the label `label`.
     */
    get label() {
        return this.fetch().label;
    }
    /**
     * Returns the child node that has the label `colon`.
     */
    get colon() {
        return this.fetch().colon;
    }
}
/**
 * This node represents a `YulFunctionCallExpression` nonterminal, with the following structure:
 *
 * ```ebnf
 * (* Postfix unary operator *)
 * YulFunctionCallExpression = (* operand: *) YulExpression
 *                             (* open_paren: *) OPEN_PAREN
 *                             (* arguments: *) YulArguments
 *                             (* close_paren: *) CLOSE_PAREN;
 * ```
 */
export class YulFunctionCallExpression {
    /**
     * Constructs a new AST node of type `YulFunctionCallExpression`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `YulFunctionCallExpression`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const [$operand, $openParen, $arguments, $closeParen] = wasm.ast.Selectors.sequence(this.cst);
            return {
                operand: new YulExpression($operand),
                openParen: $openParen,
                arguments: new YulArguments($arguments),
                closeParen: $closeParen,
            };
        });
        assertKind(this.cst.kind, NonterminalKind.YulFunctionCallExpression);
    }
    /**
     * Returns the child node that has the label `operand`.
     */
    get operand() {
        return this.fetch().operand;
    }
    /**
     * Returns the child node that has the label `open_paren`.
     */
    get openParen() {
        return this.fetch().openParen;
    }
    /**
     * Returns the child node that has the label `arguments`.
     */
    get arguments() {
        return this.fetch().arguments;
    }
    /**
     * Returns the child node that has the label `close_paren`.
     */
    get closeParen() {
        return this.fetch().closeParen;
    }
}
//
// Choices:
//
/**
 * This node represents a `SourceUnitMember` nonterminal, with the following structure:
 *
 * ```ebnf
 * SourceUnitMember = (* variant: *) PragmaDirective
 *                  | (* variant: *) ImportDirective
 *                  | (* variant: *) ContractDefinition
 *                  | (* variant: *) InterfaceDefinition
 *                  | (* variant: *) LibraryDefinition
 *                  | (* variant: *) StructDefinition (* Introduced in 0.6.0 *)
 *                  | (* variant: *) EnumDefinition (* Introduced in 0.6.0 *)
 *                  | (* variant: *) FunctionDefinition (* Introduced in 0.7.1 *)
 *                  | (* variant: *) ErrorDefinition (* Introduced in 0.8.4 *)
 *                  | (* variant: *) UserDefinedValueTypeDefinition (* Introduced in 0.8.8 *)
 *                  | (* variant: *) UsingDirective (* Introduced in 0.8.13 *)
 *                  | (* variant: *) EventDefinition (* Introduced in 0.8.22 *)
 *                  | (* variant: *) ConstantDefinition; (* Introduced in 0.7.4 *)
 * ```
 */
export class SourceUnitMember {
    /**
     * Constructs a new AST node of type `SourceUnitMember`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `SourceUnitMember`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const variant = wasm.ast.Selectors.choice(this.cst);
            switch (variant.kind) {
                case NonterminalKind.PragmaDirective:
                    return new PragmaDirective(variant);
                case NonterminalKind.ImportDirective:
                    return new ImportDirective(variant);
                case NonterminalKind.ContractDefinition:
                    return new ContractDefinition(variant);
                case NonterminalKind.InterfaceDefinition:
                    return new InterfaceDefinition(variant);
                case NonterminalKind.LibraryDefinition:
                    return new LibraryDefinition(variant);
                case NonterminalKind.StructDefinition:
                    return new StructDefinition(variant);
                case NonterminalKind.EnumDefinition:
                    return new EnumDefinition(variant);
                case NonterminalKind.FunctionDefinition:
                    return new FunctionDefinition(variant);
                case NonterminalKind.ErrorDefinition:
                    return new ErrorDefinition(variant);
                case NonterminalKind.UserDefinedValueTypeDefinition:
                    return new UserDefinedValueTypeDefinition(variant);
                case NonterminalKind.UsingDirective:
                    return new UsingDirective(variant);
                case NonterminalKind.EventDefinition:
                    return new EventDefinition(variant);
                case NonterminalKind.ConstantDefinition:
                    return new ConstantDefinition(variant);
                default:
                    throw new Error(`Unexpected variant: '${variant.kind}'.`);
            }
        });
        assertKind(this.cst.kind, NonterminalKind.SourceUnitMember);
    }
    /**
     * Returns the child node that has the label `variant`.
     */
    get variant() {
        return this.fetch();
    }
}
/**
 * This node represents a `Pragma` nonterminal, with the following structure:
 *
 * ```ebnf
 * Pragma = (* variant: *) VersionPragma
 *        | (* variant: *) AbicoderPragma (* Introduced in 0.7.5 *)
 *        | (* variant: *) ExperimentalPragma; (* Introduced in 0.4.16 *)
 * ```
 */
export class Pragma {
    /**
     * Constructs a new AST node of type `Pragma`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `Pragma`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const variant = wasm.ast.Selectors.choice(this.cst);
            switch (variant.kind) {
                case NonterminalKind.VersionPragma:
                    return new VersionPragma(variant);
                case NonterminalKind.AbicoderPragma:
                    return new AbicoderPragma(variant);
                case NonterminalKind.ExperimentalPragma:
                    return new ExperimentalPragma(variant);
                default:
                    throw new Error(`Unexpected variant: '${variant.kind}'.`);
            }
        });
        assertKind(this.cst.kind, NonterminalKind.Pragma);
    }
    /**
     * Returns the child node that has the label `variant`.
     */
    get variant() {
        return this.fetch();
    }
}
/**
 * This node represents a `AbicoderVersion` nonterminal, with the following structure:
 *
 * ```ebnf
 * (* Introduced in 0.7.5 *)
 * AbicoderVersion = (* variant: *) ABICODER_V1_KEYWORD
 *                 | (* variant: *) ABICODER_V2_KEYWORD;
 * ```
 */
export class AbicoderVersion {
    /**
     * Constructs a new AST node of type `AbicoderVersion`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `AbicoderVersion`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const variant = wasm.ast.Selectors.choice(this.cst);
            return variant;
        });
        assertKind(this.cst.kind, NonterminalKind.AbicoderVersion);
    }
    /**
     * Returns the child node that has the label `variant`.
     */
    get variant() {
        return this.fetch();
    }
}
/**
 * This node represents a `ExperimentalFeature` nonterminal, with the following structure:
 *
 * ```ebnf
 * (* Introduced in 0.4.16 *)
 * ExperimentalFeature = (* variant: *) ABI_ENCODER_V2_KEYWORD
 *                     | (* variant: *) SMT_CHECKER_KEYWORD
 *                     | (* variant: *) StringLiteral;
 * ```
 */
export class ExperimentalFeature {
    /**
     * Constructs a new AST node of type `ExperimentalFeature`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `ExperimentalFeature`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const variant = wasm.ast.Selectors.choice(this.cst);
            if (variant.isTerminalNode()) {
                return variant;
            }
            switch (variant.kind) {
                case NonterminalKind.StringLiteral:
                    return new StringLiteral(variant);
                default:
                    throw new Error(`Unexpected variant: '${variant.kind}'.`);
            }
        });
        assertKind(this.cst.kind, NonterminalKind.ExperimentalFeature);
    }
    /**
     * Returns the child node that has the label `variant`.
     */
    get variant() {
        return this.fetch();
    }
}
/**
 * This node represents a `VersionExpression` nonterminal, with the following structure:
 *
 * ```ebnf
 * VersionExpression = (* variant: *) VersionRange
 *                   | (* variant: *) VersionTerm;
 * ```
 */
export class VersionExpression {
    /**
     * Constructs a new AST node of type `VersionExpression`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `VersionExpression`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const variant = wasm.ast.Selectors.choice(this.cst);
            switch (variant.kind) {
                case NonterminalKind.VersionRange:
                    return new VersionRange(variant);
                case NonterminalKind.VersionTerm:
                    return new VersionTerm(variant);
                default:
                    throw new Error(`Unexpected variant: '${variant.kind}'.`);
            }
        });
        assertKind(this.cst.kind, NonterminalKind.VersionExpression);
    }
    /**
     * Returns the child node that has the label `variant`.
     */
    get variant() {
        return this.fetch();
    }
}
/**
 * This node represents a `VersionOperator` nonterminal, with the following structure:
 *
 * ```ebnf
 * VersionOperator = (* variant: *) CARET
 *                 | (* variant: *) TILDE
 *                 | (* variant: *) EQUAL
 *                 | (* variant: *) LESS_THAN
 *                 | (* variant: *) GREATER_THAN
 *                 | (* variant: *) LESS_THAN_EQUAL
 *                 | (* variant: *) GREATER_THAN_EQUAL;
 * ```
 */
export class VersionOperator {
    /**
     * Constructs a new AST node of type `VersionOperator`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `VersionOperator`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const variant = wasm.ast.Selectors.choice(this.cst);
            return variant;
        });
        assertKind(this.cst.kind, NonterminalKind.VersionOperator);
    }
    /**
     * Returns the child node that has the label `variant`.
     */
    get variant() {
        return this.fetch();
    }
}
/**
 * This node represents a `VersionLiteral` nonterminal, with the following structure:
 *
 * ```ebnf
 * VersionLiteral = (* variant: *) SimpleVersionLiteral
 *                | (* variant: *) SINGLE_QUOTED_VERSION_LITERAL
 *                | (* variant: *) DOUBLE_QUOTED_VERSION_LITERAL;
 * ```
 */
export class VersionLiteral {
    /**
     * Constructs a new AST node of type `VersionLiteral`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `VersionLiteral`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const variant = wasm.ast.Selectors.choice(this.cst);
            if (variant.isTerminalNode()) {
                return variant;
            }
            switch (variant.kind) {
                case NonterminalKind.SimpleVersionLiteral:
                    return new SimpleVersionLiteral(variant);
                default:
                    throw new Error(`Unexpected variant: '${variant.kind}'.`);
            }
        });
        assertKind(this.cst.kind, NonterminalKind.VersionLiteral);
    }
    /**
     * Returns the child node that has the label `variant`.
     */
    get variant() {
        return this.fetch();
    }
}
/**
 * This node represents a `ImportClause` nonterminal, with the following structure:
 *
 * ```ebnf
 * ImportClause = (* variant: *) PathImport
 *              | (* variant: *) NamedImport
 *              | (* variant: *) ImportDeconstruction;
 * ```
 */
export class ImportClause {
    /**
     * Constructs a new AST node of type `ImportClause`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `ImportClause`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const variant = wasm.ast.Selectors.choice(this.cst);
            switch (variant.kind) {
                case NonterminalKind.PathImport:
                    return new PathImport(variant);
                case NonterminalKind.NamedImport:
                    return new NamedImport(variant);
                case NonterminalKind.ImportDeconstruction:
                    return new ImportDeconstruction(variant);
                default:
                    throw new Error(`Unexpected variant: '${variant.kind}'.`);
            }
        });
        assertKind(this.cst.kind, NonterminalKind.ImportClause);
    }
    /**
     * Returns the child node that has the label `variant`.
     */
    get variant() {
        return this.fetch();
    }
}
/**
 * This node represents a `UsingClause` nonterminal, with the following structure:
 *
 * ```ebnf
 * UsingClause = (* variant: *) IdentifierPath
 *             | (* variant: *) UsingDeconstruction; (* Introduced in 0.8.13 *)
 * ```
 */
export class UsingClause {
    /**
     * Constructs a new AST node of type `UsingClause`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `UsingClause`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const variant = wasm.ast.Selectors.choice(this.cst);
            switch (variant.kind) {
                case NonterminalKind.IdentifierPath:
                    return new IdentifierPath(variant);
                case NonterminalKind.UsingDeconstruction:
                    return new UsingDeconstruction(variant);
                default:
                    throw new Error(`Unexpected variant: '${variant.kind}'.`);
            }
        });
        assertKind(this.cst.kind, NonterminalKind.UsingClause);
    }
    /**
     * Returns the child node that has the label `variant`.
     */
    get variant() {
        return this.fetch();
    }
}
/**
 * This node represents a `UsingOperator` nonterminal, with the following structure:
 *
 * ```ebnf
 * (* Introduced in 0.8.19 *)
 * UsingOperator = (* variant: *) AMPERSAND
 *               | (* variant: *) ASTERISK
 *               | (* variant: *) BANG_EQUAL
 *               | (* variant: *) BAR
 *               | (* variant: *) CARET
 *               | (* variant: *) EQUAL_EQUAL
 *               | (* variant: *) GREATER_THAN
 *               | (* variant: *) GREATER_THAN_EQUAL
 *               | (* variant: *) LESS_THAN
 *               | (* variant: *) LESS_THAN_EQUAL
 *               | (* variant: *) MINUS
 *               | (* variant: *) PERCENT
 *               | (* variant: *) PLUS
 *               | (* variant: *) SLASH
 *               | (* variant: *) TILDE;
 * ```
 */
export class UsingOperator {
    /**
     * Constructs a new AST node of type `UsingOperator`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `UsingOperator`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const variant = wasm.ast.Selectors.choice(this.cst);
            return variant;
        });
        assertKind(this.cst.kind, NonterminalKind.UsingOperator);
    }
    /**
     * Returns the child node that has the label `variant`.
     */
    get variant() {
        return this.fetch();
    }
}
/**
 * This node represents a `UsingTarget` nonterminal, with the following structure:
 *
 * ```ebnf
 * UsingTarget = (* variant: *) TypeName
 *             | (* variant: *) ASTERISK;
 * ```
 */
export class UsingTarget {
    /**
     * Constructs a new AST node of type `UsingTarget`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `UsingTarget`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const variant = wasm.ast.Selectors.choice(this.cst);
            if (variant.isTerminalNode()) {
                return variant;
            }
            switch (variant.kind) {
                case NonterminalKind.TypeName:
                    return new TypeName(variant);
                default:
                    throw new Error(`Unexpected variant: '${variant.kind}'.`);
            }
        });
        assertKind(this.cst.kind, NonterminalKind.UsingTarget);
    }
    /**
     * Returns the child node that has the label `variant`.
     */
    get variant() {
        return this.fetch();
    }
}
/**
 * This node represents a `ContractSpecifier` nonterminal, with the following structure:
 *
 * ```ebnf
 * ContractSpecifier = (* variant: *) InheritanceSpecifier
 *                   | (* variant: *) StorageLayoutSpecifier; (* Introduced in 0.8.29 *)
 * ```
 */
export class ContractSpecifier {
    /**
     * Constructs a new AST node of type `ContractSpecifier`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `ContractSpecifier`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const variant = wasm.ast.Selectors.choice(this.cst);
            switch (variant.kind) {
                case NonterminalKind.InheritanceSpecifier:
                    return new InheritanceSpecifier(variant);
                case NonterminalKind.StorageLayoutSpecifier:
                    return new StorageLayoutSpecifier(variant);
                default:
                    throw new Error(`Unexpected variant: '${variant.kind}'.`);
            }
        });
        assertKind(this.cst.kind, NonterminalKind.ContractSpecifier);
    }
    /**
     * Returns the child node that has the label `variant`.
     */
    get variant() {
        return this.fetch();
    }
}
/**
 * This node represents a `ContractMember` nonterminal, with the following structure:
 *
 * ```ebnf
 * ContractMember = (* variant: *) UsingDirective
 *                | (* variant: *) FunctionDefinition
 *                | (* variant: *) ConstructorDefinition (* Introduced in 0.4.22 *)
 *                | (* variant: *) ReceiveFunctionDefinition (* Introduced in 0.6.0 *)
 *                | (* variant: *) FallbackFunctionDefinition (* Introduced in 0.6.0 *)
 *                | (* variant: *) UnnamedFunctionDefinition (* Deprecated in 0.6.0 *)
 *                | (* variant: *) ModifierDefinition
 *                | (* variant: *) StructDefinition
 *                | (* variant: *) EnumDefinition
 *                | (* variant: *) EventDefinition
 *                | (* variant: *) ErrorDefinition (* Introduced in 0.8.4 *)
 *                | (* variant: *) UserDefinedValueTypeDefinition (* Introduced in 0.8.8 *)
 *                | (* variant: *) StateVariableDefinition;
 * ```
 */
export class ContractMember {
    /**
     * Constructs a new AST node of type `ContractMember`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `ContractMember`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const variant = wasm.ast.Selectors.choice(this.cst);
            switch (variant.kind) {
                case NonterminalKind.UsingDirective:
                    return new UsingDirective(variant);
                case NonterminalKind.FunctionDefinition:
                    return new FunctionDefinition(variant);
                case NonterminalKind.ConstructorDefinition:
                    return new ConstructorDefinition(variant);
                case NonterminalKind.ReceiveFunctionDefinition:
                    return new ReceiveFunctionDefinition(variant);
                case NonterminalKind.FallbackFunctionDefinition:
                    return new FallbackFunctionDefinition(variant);
                case NonterminalKind.UnnamedFunctionDefinition:
                    return new UnnamedFunctionDefinition(variant);
                case NonterminalKind.ModifierDefinition:
                    return new ModifierDefinition(variant);
                case NonterminalKind.StructDefinition:
                    return new StructDefinition(variant);
                case NonterminalKind.EnumDefinition:
                    return new EnumDefinition(variant);
                case NonterminalKind.EventDefinition:
                    return new EventDefinition(variant);
                case NonterminalKind.ErrorDefinition:
                    return new ErrorDefinition(variant);
                case NonterminalKind.UserDefinedValueTypeDefinition:
                    return new UserDefinedValueTypeDefinition(variant);
                case NonterminalKind.StateVariableDefinition:
                    return new StateVariableDefinition(variant);
                default:
                    throw new Error(`Unexpected variant: '${variant.kind}'.`);
            }
        });
        assertKind(this.cst.kind, NonterminalKind.ContractMember);
    }
    /**
     * Returns the child node that has the label `variant`.
     */
    get variant() {
        return this.fetch();
    }
}
/**
 * This node represents a `StateVariableAttribute` nonterminal, with the following structure:
 *
 * ```ebnf
 * StateVariableAttribute = (* variant: *) OverrideSpecifier (* Introduced in 0.6.0 *)
 *                        | (* variant: *) CONSTANT_KEYWORD
 *                        | (* variant: *) INTERNAL_KEYWORD
 *                        | (* variant: *) PRIVATE_KEYWORD
 *                        | (* variant: *) PUBLIC_KEYWORD
 *                        | (* variant: *) IMMUTABLE_KEYWORD (* Introduced in 0.6.5 *)
 *                        | (* variant: *) TRANSIENT_KEYWORD; (* Introduced in 0.8.27 *)
 * ```
 */
export class StateVariableAttribute {
    /**
     * Constructs a new AST node of type `StateVariableAttribute`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `StateVariableAttribute`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const variant = wasm.ast.Selectors.choice(this.cst);
            if (variant.isTerminalNode()) {
                return variant;
            }
            switch (variant.kind) {
                case NonterminalKind.OverrideSpecifier:
                    return new OverrideSpecifier(variant);
                default:
                    throw new Error(`Unexpected variant: '${variant.kind}'.`);
            }
        });
        assertKind(this.cst.kind, NonterminalKind.StateVariableAttribute);
    }
    /**
     * Returns the child node that has the label `variant`.
     */
    get variant() {
        return this.fetch();
    }
}
/**
 * This node represents a `FunctionName` nonterminal, with the following structure:
 *
 * ```ebnf
 * FunctionName = (* variant: *) IDENTIFIER
 *              | (* variant: *) FALLBACK_KEYWORD
 *              | (* variant: *) RECEIVE_KEYWORD;
 * ```
 */
export class FunctionName {
    /**
     * Constructs a new AST node of type `FunctionName`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `FunctionName`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const variant = wasm.ast.Selectors.choice(this.cst);
            return variant;
        });
        assertKind(this.cst.kind, NonterminalKind.FunctionName);
    }
    /**
     * Returns the child node that has the label `variant`.
     */
    get variant() {
        return this.fetch();
    }
}
/**
 * This node represents a `FunctionAttribute` nonterminal, with the following structure:
 *
 * ```ebnf
 * FunctionAttribute = (* variant: *) ModifierInvocation
 *                   | (* variant: *) OverrideSpecifier (* Introduced in 0.6.0 *)
 *                   | (* variant: *) CONSTANT_KEYWORD (* Deprecated in 0.5.0 *)
 *                   | (* variant: *) EXTERNAL_KEYWORD
 *                   | (* variant: *) INTERNAL_KEYWORD
 *                   | (* variant: *) PAYABLE_KEYWORD
 *                   | (* variant: *) PRIVATE_KEYWORD
 *                   | (* variant: *) PUBLIC_KEYWORD
 *                   | (* variant: *) PURE_KEYWORD (* Introduced in 0.4.16 *)
 *                   | (* variant: *) VIEW_KEYWORD (* Introduced in 0.4.16 *)
 *                   | (* variant: *) VIRTUAL_KEYWORD; (* Introduced in 0.6.0 *)
 * ```
 */
export class FunctionAttribute {
    /**
     * Constructs a new AST node of type `FunctionAttribute`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `FunctionAttribute`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const variant = wasm.ast.Selectors.choice(this.cst);
            if (variant.isTerminalNode()) {
                return variant;
            }
            switch (variant.kind) {
                case NonterminalKind.ModifierInvocation:
                    return new ModifierInvocation(variant);
                case NonterminalKind.OverrideSpecifier:
                    return new OverrideSpecifier(variant);
                default:
                    throw new Error(`Unexpected variant: '${variant.kind}'.`);
            }
        });
        assertKind(this.cst.kind, NonterminalKind.FunctionAttribute);
    }
    /**
     * Returns the child node that has the label `variant`.
     */
    get variant() {
        return this.fetch();
    }
}
/**
 * This node represents a `FunctionBody` nonterminal, with the following structure:
 *
 * ```ebnf
 * FunctionBody = (* variant: *) Block
 *              | (* variant: *) SEMICOLON;
 * ```
 */
export class FunctionBody {
    /**
     * Constructs a new AST node of type `FunctionBody`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `FunctionBody`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const variant = wasm.ast.Selectors.choice(this.cst);
            if (variant.isTerminalNode()) {
                return variant;
            }
            switch (variant.kind) {
                case NonterminalKind.Block:
                    return new Block(variant);
                default:
                    throw new Error(`Unexpected variant: '${variant.kind}'.`);
            }
        });
        assertKind(this.cst.kind, NonterminalKind.FunctionBody);
    }
    /**
     * Returns the child node that has the label `variant`.
     */
    get variant() {
        return this.fetch();
    }
}
/**
 * This node represents a `ConstructorAttribute` nonterminal, with the following structure:
 *
 * ```ebnf
 * (* Introduced in 0.4.22 *)
 * ConstructorAttribute = (* variant: *) ModifierInvocation
 *                      | (* variant: *) INTERNAL_KEYWORD
 *                      | (* variant: *) OVERRIDE_KEYWORD (* Introduced in 0.6.0 and deprecated in 0.6.7. *)
 *                      | (* variant: *) PAYABLE_KEYWORD
 *                      | (* variant: *) PUBLIC_KEYWORD
 *                      | (* variant: *) VIRTUAL_KEYWORD; (* Introduced in 0.6.0 and deprecated in 0.6.7. *)
 * ```
 */
export class ConstructorAttribute {
    /**
     * Constructs a new AST node of type `ConstructorAttribute`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `ConstructorAttribute`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const variant = wasm.ast.Selectors.choice(this.cst);
            if (variant.isTerminalNode()) {
                return variant;
            }
            switch (variant.kind) {
                case NonterminalKind.ModifierInvocation:
                    return new ModifierInvocation(variant);
                default:
                    throw new Error(`Unexpected variant: '${variant.kind}'.`);
            }
        });
        assertKind(this.cst.kind, NonterminalKind.ConstructorAttribute);
    }
    /**
     * Returns the child node that has the label `variant`.
     */
    get variant() {
        return this.fetch();
    }
}
/**
 * This node represents a `UnnamedFunctionAttribute` nonterminal, with the following structure:
 *
 * ```ebnf
 * (* Deprecated in 0.6.0 *)
 * UnnamedFunctionAttribute = (* variant: *) ModifierInvocation
 *                          | (* variant: *) EXTERNAL_KEYWORD
 *                          | (* variant: *) INTERNAL_KEYWORD (* Deprecated in 0.5.0 *)
 *                          | (* variant: *) PAYABLE_KEYWORD
 *                          | (* variant: *) PRIVATE_KEYWORD (* Deprecated in 0.5.0 *)
 *                          | (* variant: *) PUBLIC_KEYWORD; (* Deprecated in 0.5.0 *)
 * ```
 */
export class UnnamedFunctionAttribute {
    /**
     * Constructs a new AST node of type `UnnamedFunctionAttribute`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `UnnamedFunctionAttribute`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const variant = wasm.ast.Selectors.choice(this.cst);
            if (variant.isTerminalNode()) {
                return variant;
            }
            switch (variant.kind) {
                case NonterminalKind.ModifierInvocation:
                    return new ModifierInvocation(variant);
                default:
                    throw new Error(`Unexpected variant: '${variant.kind}'.`);
            }
        });
        assertKind(this.cst.kind, NonterminalKind.UnnamedFunctionAttribute);
    }
    /**
     * Returns the child node that has the label `variant`.
     */
    get variant() {
        return this.fetch();
    }
}
/**
 * This node represents a `FallbackFunctionAttribute` nonterminal, with the following structure:
 *
 * ```ebnf
 * (* Introduced in 0.6.0 *)
 * FallbackFunctionAttribute = (* variant: *) ModifierInvocation
 *                           | (* variant: *) OverrideSpecifier
 *                           | (* variant: *) EXTERNAL_KEYWORD
 *                           | (* variant: *) PAYABLE_KEYWORD
 *                           | (* variant: *) PURE_KEYWORD
 *                           | (* variant: *) VIEW_KEYWORD
 *                           | (* variant: *) VIRTUAL_KEYWORD;
 * ```
 */
export class FallbackFunctionAttribute {
    /**
     * Constructs a new AST node of type `FallbackFunctionAttribute`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `FallbackFunctionAttribute`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const variant = wasm.ast.Selectors.choice(this.cst);
            if (variant.isTerminalNode()) {
                return variant;
            }
            switch (variant.kind) {
                case NonterminalKind.ModifierInvocation:
                    return new ModifierInvocation(variant);
                case NonterminalKind.OverrideSpecifier:
                    return new OverrideSpecifier(variant);
                default:
                    throw new Error(`Unexpected variant: '${variant.kind}'.`);
            }
        });
        assertKind(this.cst.kind, NonterminalKind.FallbackFunctionAttribute);
    }
    /**
     * Returns the child node that has the label `variant`.
     */
    get variant() {
        return this.fetch();
    }
}
/**
 * This node represents a `ReceiveFunctionAttribute` nonterminal, with the following structure:
 *
 * ```ebnf
 * (* Introduced in 0.6.0 *)
 * ReceiveFunctionAttribute = (* variant: *) ModifierInvocation
 *                          | (* variant: *) OverrideSpecifier
 *                          | (* variant: *) EXTERNAL_KEYWORD
 *                          | (* variant: *) PAYABLE_KEYWORD
 *                          | (* variant: *) VIRTUAL_KEYWORD;
 * ```
 */
export class ReceiveFunctionAttribute {
    /**
     * Constructs a new AST node of type `ReceiveFunctionAttribute`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `ReceiveFunctionAttribute`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const variant = wasm.ast.Selectors.choice(this.cst);
            if (variant.isTerminalNode()) {
                return variant;
            }
            switch (variant.kind) {
                case NonterminalKind.ModifierInvocation:
                    return new ModifierInvocation(variant);
                case NonterminalKind.OverrideSpecifier:
                    return new OverrideSpecifier(variant);
                default:
                    throw new Error(`Unexpected variant: '${variant.kind}'.`);
            }
        });
        assertKind(this.cst.kind, NonterminalKind.ReceiveFunctionAttribute);
    }
    /**
     * Returns the child node that has the label `variant`.
     */
    get variant() {
        return this.fetch();
    }
}
/**
 * This node represents a `ModifierAttribute` nonterminal, with the following structure:
 *
 * ```ebnf
 * ModifierAttribute = (* variant: *) OverrideSpecifier (* Introduced in 0.6.0 *)
 *                   | (* variant: *) VIRTUAL_KEYWORD; (* Introduced in 0.6.0 *)
 * ```
 */
export class ModifierAttribute {
    /**
     * Constructs a new AST node of type `ModifierAttribute`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `ModifierAttribute`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const variant = wasm.ast.Selectors.choice(this.cst);
            if (variant.isTerminalNode()) {
                return variant;
            }
            switch (variant.kind) {
                case NonterminalKind.OverrideSpecifier:
                    return new OverrideSpecifier(variant);
                default:
                    throw new Error(`Unexpected variant: '${variant.kind}'.`);
            }
        });
        assertKind(this.cst.kind, NonterminalKind.ModifierAttribute);
    }
    /**
     * Returns the child node that has the label `variant`.
     */
    get variant() {
        return this.fetch();
    }
}
/**
 * This node represents a `TypeName` nonterminal, with the following structure:
 *
 * ```ebnf
 * TypeName = (* variant: *) ArrayTypeName
 *          | (* variant: *) FunctionType
 *          | (* variant: *) MappingType
 *          | (* variant: *) ElementaryType
 *          | (* variant: *) IdentifierPath;
 * ```
 */
export class TypeName {
    /**
     * Constructs a new AST node of type `TypeName`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `TypeName`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const variant = wasm.ast.Selectors.choice(this.cst);
            switch (variant.kind) {
                case NonterminalKind.ArrayTypeName:
                    return new ArrayTypeName(variant);
                case NonterminalKind.FunctionType:
                    return new FunctionType(variant);
                case NonterminalKind.MappingType:
                    return new MappingType(variant);
                case NonterminalKind.ElementaryType:
                    return new ElementaryType(variant);
                case NonterminalKind.IdentifierPath:
                    return new IdentifierPath(variant);
                default:
                    throw new Error(`Unexpected variant: '${variant.kind}'.`);
            }
        });
        assertKind(this.cst.kind, NonterminalKind.TypeName);
    }
    /**
     * Returns the child node that has the label `variant`.
     */
    get variant() {
        return this.fetch();
    }
}
/**
 * This node represents a `FunctionTypeAttribute` nonterminal, with the following structure:
 *
 * ```ebnf
 * FunctionTypeAttribute = (* variant: *) INTERNAL_KEYWORD
 *                       | (* variant: *) EXTERNAL_KEYWORD
 *                       | (* variant: *) PRIVATE_KEYWORD
 *                       | (* variant: *) PUBLIC_KEYWORD
 *                       | (* variant: *) CONSTANT_KEYWORD (* Deprecated in 0.5.0 *)
 *                       | (* variant: *) PURE_KEYWORD (* Introduced in 0.4.16 *)
 *                       | (* variant: *) VIEW_KEYWORD (* Introduced in 0.4.16 *)
 *                       | (* variant: *) PAYABLE_KEYWORD;
 * ```
 */
export class FunctionTypeAttribute {
    /**
     * Constructs a new AST node of type `FunctionTypeAttribute`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `FunctionTypeAttribute`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const variant = wasm.ast.Selectors.choice(this.cst);
            return variant;
        });
        assertKind(this.cst.kind, NonterminalKind.FunctionTypeAttribute);
    }
    /**
     * Returns the child node that has the label `variant`.
     */
    get variant() {
        return this.fetch();
    }
}
/**
 * This node represents a `MappingKeyType` nonterminal, with the following structure:
 *
 * ```ebnf
 * MappingKeyType = (* variant: *) ElementaryType
 *                | (* variant: *) IdentifierPath;
 * ```
 */
export class MappingKeyType {
    /**
     * Constructs a new AST node of type `MappingKeyType`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `MappingKeyType`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const variant = wasm.ast.Selectors.choice(this.cst);
            switch (variant.kind) {
                case NonterminalKind.ElementaryType:
                    return new ElementaryType(variant);
                case NonterminalKind.IdentifierPath:
                    return new IdentifierPath(variant);
                default:
                    throw new Error(`Unexpected variant: '${variant.kind}'.`);
            }
        });
        assertKind(this.cst.kind, NonterminalKind.MappingKeyType);
    }
    /**
     * Returns the child node that has the label `variant`.
     */
    get variant() {
        return this.fetch();
    }
}
/**
 * This node represents a `ElementaryType` nonterminal, with the following structure:
 *
 * ```ebnf
 * ElementaryType = (* variant: *) BOOL_KEYWORD
 *                | (* variant: *) BYTE_KEYWORD (* Deprecated in 0.8.0 *)
 *                | (* variant: *) STRING_KEYWORD
 *                | (* variant: *) AddressType
 *                | (* variant: *) BYTES_KEYWORD
 *                | (* variant: *) INT_KEYWORD
 *                | (* variant: *) UINT_KEYWORD
 *                | (* variant: *) FIXED_KEYWORD
 *                | (* variant: *) UFIXED_KEYWORD;
 * ```
 */
export class ElementaryType {
    /**
     * Constructs a new AST node of type `ElementaryType`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `ElementaryType`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const variant = wasm.ast.Selectors.choice(this.cst);
            if (variant.isTerminalNode()) {
                return variant;
            }
            switch (variant.kind) {
                case NonterminalKind.AddressType:
                    return new AddressType(variant);
                default:
                    throw new Error(`Unexpected variant: '${variant.kind}'.`);
            }
        });
        assertKind(this.cst.kind, NonterminalKind.ElementaryType);
    }
    /**
     * Returns the child node that has the label `variant`.
     */
    get variant() {
        return this.fetch();
    }
}
/**
 * This node represents a `Statement` nonterminal, with the following structure:
 *
 * ```ebnf
 * Statement = (* variant: *) IfStatement
 *           | (* variant: *) ForStatement
 *           | (* variant: *) WhileStatement
 *           | (* variant: *) DoWhileStatement
 *           | (* variant: *) ContinueStatement
 *           | (* variant: *) BreakStatement
 *           | (* variant: *) ReturnStatement
 *           | (* variant: *) ThrowStatement (* Deprecated in 0.5.0 *)
 *           | (* variant: *) EmitStatement (* Introduced in 0.4.21 *)
 *           | (* variant: *) TryStatement (* Introduced in 0.6.0 *)
 *           | (* variant: *) RevertStatement (* Introduced in 0.8.4 *)
 *           | (* variant: *) AssemblyStatement
 *           | (* variant: *) Block
 *           | (* variant: *) UncheckedBlock (* Introduced in 0.8.0 *)
 *           | (* variant: *) TupleDeconstructionStatement
 *           | (* variant: *) VariableDeclarationStatement
 *           | (* variant: *) ExpressionStatement;
 * ```
 */
export class Statement {
    /**
     * Constructs a new AST node of type `Statement`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `Statement`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const variant = wasm.ast.Selectors.choice(this.cst);
            switch (variant.kind) {
                case NonterminalKind.IfStatement:
                    return new IfStatement(variant);
                case NonterminalKind.ForStatement:
                    return new ForStatement(variant);
                case NonterminalKind.WhileStatement:
                    return new WhileStatement(variant);
                case NonterminalKind.DoWhileStatement:
                    return new DoWhileStatement(variant);
                case NonterminalKind.ContinueStatement:
                    return new ContinueStatement(variant);
                case NonterminalKind.BreakStatement:
                    return new BreakStatement(variant);
                case NonterminalKind.ReturnStatement:
                    return new ReturnStatement(variant);
                case NonterminalKind.ThrowStatement:
                    return new ThrowStatement(variant);
                case NonterminalKind.EmitStatement:
                    return new EmitStatement(variant);
                case NonterminalKind.TryStatement:
                    return new TryStatement(variant);
                case NonterminalKind.RevertStatement:
                    return new RevertStatement(variant);
                case NonterminalKind.AssemblyStatement:
                    return new AssemblyStatement(variant);
                case NonterminalKind.Block:
                    return new Block(variant);
                case NonterminalKind.UncheckedBlock:
                    return new UncheckedBlock(variant);
                case NonterminalKind.TupleDeconstructionStatement:
                    return new TupleDeconstructionStatement(variant);
                case NonterminalKind.VariableDeclarationStatement:
                    return new VariableDeclarationStatement(variant);
                case NonterminalKind.ExpressionStatement:
                    return new ExpressionStatement(variant);
                default:
                    throw new Error(`Unexpected variant: '${variant.kind}'.`);
            }
        });
        assertKind(this.cst.kind, NonterminalKind.Statement);
    }
    /**
     * Returns the child node that has the label `variant`.
     */
    get variant() {
        return this.fetch();
    }
}
/**
 * This node represents a `TupleMember` nonterminal, with the following structure:
 *
 * ```ebnf
 * TupleMember = (* variant: *) TypedTupleMember
 *             | (* variant: *) UntypedTupleMember;
 * ```
 */
export class TupleMember {
    /**
     * Constructs a new AST node of type `TupleMember`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `TupleMember`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const variant = wasm.ast.Selectors.choice(this.cst);
            switch (variant.kind) {
                case NonterminalKind.TypedTupleMember:
                    return new TypedTupleMember(variant);
                case NonterminalKind.UntypedTupleMember:
                    return new UntypedTupleMember(variant);
                default:
                    throw new Error(`Unexpected variant: '${variant.kind}'.`);
            }
        });
        assertKind(this.cst.kind, NonterminalKind.TupleMember);
    }
    /**
     * Returns the child node that has the label `variant`.
     */
    get variant() {
        return this.fetch();
    }
}
/**
 * This node represents a `VariableDeclarationType` nonterminal, with the following structure:
 *
 * ```ebnf
 * VariableDeclarationType = (* variant: *) TypeName
 *                         | (* variant: *) VAR_KEYWORD; (* Deprecated in 0.5.0 *)
 * ```
 */
export class VariableDeclarationType {
    /**
     * Constructs a new AST node of type `VariableDeclarationType`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `VariableDeclarationType`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const variant = wasm.ast.Selectors.choice(this.cst);
            if (variant.isTerminalNode()) {
                return variant;
            }
            switch (variant.kind) {
                case NonterminalKind.TypeName:
                    return new TypeName(variant);
                default:
                    throw new Error(`Unexpected variant: '${variant.kind}'.`);
            }
        });
        assertKind(this.cst.kind, NonterminalKind.VariableDeclarationType);
    }
    /**
     * Returns the child node that has the label `variant`.
     */
    get variant() {
        return this.fetch();
    }
}
/**
 * This node represents a `StorageLocation` nonterminal, with the following structure:
 *
 * ```ebnf
 * StorageLocation = (* variant: *) MEMORY_KEYWORD
 *                 | (* variant: *) STORAGE_KEYWORD
 *                 | (* variant: *) CALL_DATA_KEYWORD; (* Introduced in 0.5.0 *)
 * ```
 */
export class StorageLocation {
    /**
     * Constructs a new AST node of type `StorageLocation`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `StorageLocation`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const variant = wasm.ast.Selectors.choice(this.cst);
            return variant;
        });
        assertKind(this.cst.kind, NonterminalKind.StorageLocation);
    }
    /**
     * Returns the child node that has the label `variant`.
     */
    get variant() {
        return this.fetch();
    }
}
/**
 * This node represents a `ForStatementInitialization` nonterminal, with the following structure:
 *
 * ```ebnf
 * ForStatementInitialization = (* variant: *) TupleDeconstructionStatement
 *                            | (* variant: *) VariableDeclarationStatement
 *                            | (* variant: *) ExpressionStatement
 *                            | (* variant: *) SEMICOLON;
 * ```
 */
export class ForStatementInitialization {
    /**
     * Constructs a new AST node of type `ForStatementInitialization`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `ForStatementInitialization`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const variant = wasm.ast.Selectors.choice(this.cst);
            if (variant.isTerminalNode()) {
                return variant;
            }
            switch (variant.kind) {
                case NonterminalKind.TupleDeconstructionStatement:
                    return new TupleDeconstructionStatement(variant);
                case NonterminalKind.VariableDeclarationStatement:
                    return new VariableDeclarationStatement(variant);
                case NonterminalKind.ExpressionStatement:
                    return new ExpressionStatement(variant);
                default:
                    throw new Error(`Unexpected variant: '${variant.kind}'.`);
            }
        });
        assertKind(this.cst.kind, NonterminalKind.ForStatementInitialization);
    }
    /**
     * Returns the child node that has the label `variant`.
     */
    get variant() {
        return this.fetch();
    }
}
/**
 * This node represents a `ForStatementCondition` nonterminal, with the following structure:
 *
 * ```ebnf
 * ForStatementCondition = (* variant: *) ExpressionStatement
 *                       | (* variant: *) SEMICOLON;
 * ```
 */
export class ForStatementCondition {
    /**
     * Constructs a new AST node of type `ForStatementCondition`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `ForStatementCondition`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const variant = wasm.ast.Selectors.choice(this.cst);
            if (variant.isTerminalNode()) {
                return variant;
            }
            switch (variant.kind) {
                case NonterminalKind.ExpressionStatement:
                    return new ExpressionStatement(variant);
                default:
                    throw new Error(`Unexpected variant: '${variant.kind}'.`);
            }
        });
        assertKind(this.cst.kind, NonterminalKind.ForStatementCondition);
    }
    /**
     * Returns the child node that has the label `variant`.
     */
    get variant() {
        return this.fetch();
    }
}
/**
 * This node represents a `Expression` nonterminal, with the following structure:
 *
 * ```ebnf
 * Expression = (* variant: *) AssignmentExpression
 *            | (* variant: *) ConditionalExpression
 *            | (* variant: *) OrExpression
 *            | (* variant: *) AndExpression
 *            | (* variant: *) EqualityExpression
 *            | (* variant: *) InequalityExpression
 *            | (* variant: *) BitwiseOrExpression
 *            | (* variant: *) BitwiseXorExpression
 *            | (* variant: *) BitwiseAndExpression
 *            | (* variant: *) ShiftExpression
 *            | (* variant: *) AdditiveExpression
 *            | (* variant: *) MultiplicativeExpression
 *            | (* variant: *) ExponentiationExpression
 *            | (* variant: *) PostfixExpression
 *            | (* variant: *) PrefixExpression
 *            | (* variant: *) FunctionCallExpression
 *            | (* variant: *) CallOptionsExpression
 *            | (* variant: *) MemberAccessExpression
 *            | (* variant: *) IndexAccessExpression
 *            | (* variant: *) NewExpression
 *            | (* variant: *) TupleExpression
 *            | (* variant: *) TypeExpression (* Introduced in 0.5.3 *)
 *            | (* variant: *) ArrayExpression
 *            | (* variant: *) HexNumberExpression
 *            | (* variant: *) DecimalNumberExpression
 *            | (* variant: *) StringExpression
 *            | (* variant: *) ElementaryType
 *            | (* variant: *) PAYABLE_KEYWORD (* Introduced in 0.6.0 *)
 *            | (* variant: *) THIS_KEYWORD
 *            | (* variant: *) SUPER_KEYWORD
 *            | (* variant: *) TRUE_KEYWORD
 *            | (* variant: *) FALSE_KEYWORD
 *            | (* variant: *) IDENTIFIER;
 * ```
 */
export class Expression {
    /**
     * Constructs a new AST node of type `Expression`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `Expression`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const variant = wasm.ast.Selectors.choice(this.cst);
            if (variant.isTerminalNode()) {
                return variant;
            }
            switch (variant.kind) {
                case NonterminalKind.AssignmentExpression:
                    return new AssignmentExpression(variant);
                case NonterminalKind.ConditionalExpression:
                    return new ConditionalExpression(variant);
                case NonterminalKind.OrExpression:
                    return new OrExpression(variant);
                case NonterminalKind.AndExpression:
                    return new AndExpression(variant);
                case NonterminalKind.EqualityExpression:
                    return new EqualityExpression(variant);
                case NonterminalKind.InequalityExpression:
                    return new InequalityExpression(variant);
                case NonterminalKind.BitwiseOrExpression:
                    return new BitwiseOrExpression(variant);
                case NonterminalKind.BitwiseXorExpression:
                    return new BitwiseXorExpression(variant);
                case NonterminalKind.BitwiseAndExpression:
                    return new BitwiseAndExpression(variant);
                case NonterminalKind.ShiftExpression:
                    return new ShiftExpression(variant);
                case NonterminalKind.AdditiveExpression:
                    return new AdditiveExpression(variant);
                case NonterminalKind.MultiplicativeExpression:
                    return new MultiplicativeExpression(variant);
                case NonterminalKind.ExponentiationExpression:
                    return new ExponentiationExpression(variant);
                case NonterminalKind.PostfixExpression:
                    return new PostfixExpression(variant);
                case NonterminalKind.PrefixExpression:
                    return new PrefixExpression(variant);
                case NonterminalKind.FunctionCallExpression:
                    return new FunctionCallExpression(variant);
                case NonterminalKind.CallOptionsExpression:
                    return new CallOptionsExpression(variant);
                case NonterminalKind.MemberAccessExpression:
                    return new MemberAccessExpression(variant);
                case NonterminalKind.IndexAccessExpression:
                    return new IndexAccessExpression(variant);
                case NonterminalKind.NewExpression:
                    return new NewExpression(variant);
                case NonterminalKind.TupleExpression:
                    return new TupleExpression(variant);
                case NonterminalKind.TypeExpression:
                    return new TypeExpression(variant);
                case NonterminalKind.ArrayExpression:
                    return new ArrayExpression(variant);
                case NonterminalKind.HexNumberExpression:
                    return new HexNumberExpression(variant);
                case NonterminalKind.DecimalNumberExpression:
                    return new DecimalNumberExpression(variant);
                case NonterminalKind.StringExpression:
                    return new StringExpression(variant);
                case NonterminalKind.ElementaryType:
                    return new ElementaryType(variant);
                default:
                    throw new Error(`Unexpected variant: '${variant.kind}'.`);
            }
        });
        assertKind(this.cst.kind, NonterminalKind.Expression);
    }
    /**
     * Returns the child node that has the label `variant`.
     */
    get variant() {
        return this.fetch();
    }
}
/**
 * This node represents a `ArgumentsDeclaration` nonterminal, with the following structure:
 *
 * ```ebnf
 * ArgumentsDeclaration = (* variant: *) PositionalArgumentsDeclaration
 *                      | (* variant: *) NamedArgumentsDeclaration;
 * ```
 */
export class ArgumentsDeclaration {
    /**
     * Constructs a new AST node of type `ArgumentsDeclaration`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `ArgumentsDeclaration`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const variant = wasm.ast.Selectors.choice(this.cst);
            switch (variant.kind) {
                case NonterminalKind.PositionalArgumentsDeclaration:
                    return new PositionalArgumentsDeclaration(variant);
                case NonterminalKind.NamedArgumentsDeclaration:
                    return new NamedArgumentsDeclaration(variant);
                default:
                    throw new Error(`Unexpected variant: '${variant.kind}'.`);
            }
        });
        assertKind(this.cst.kind, NonterminalKind.ArgumentsDeclaration);
    }
    /**
     * Returns the child node that has the label `variant`.
     */
    get variant() {
        return this.fetch();
    }
}
/**
 * This node represents a `NumberUnit` nonterminal, with the following structure:
 *
 * ```ebnf
 * NumberUnit = (* variant: *) WEI_KEYWORD
 *            | (* variant: *) GWEI_KEYWORD (* Introduced in 0.6.11 *)
 *            | (* variant: *) SZABO_KEYWORD (* Deprecated in 0.7.0 *)
 *            | (* variant: *) FINNEY_KEYWORD (* Deprecated in 0.7.0 *)
 *            | (* variant: *) ETHER_KEYWORD
 *            | (* variant: *) SECONDS_KEYWORD
 *            | (* variant: *) MINUTES_KEYWORD
 *            | (* variant: *) HOURS_KEYWORD
 *            | (* variant: *) DAYS_KEYWORD
 *            | (* variant: *) WEEKS_KEYWORD
 *            | (* variant: *) YEARS_KEYWORD; (* Deprecated in 0.5.0 *)
 * ```
 */
export class NumberUnit {
    /**
     * Constructs a new AST node of type `NumberUnit`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `NumberUnit`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const variant = wasm.ast.Selectors.choice(this.cst);
            return variant;
        });
        assertKind(this.cst.kind, NonterminalKind.NumberUnit);
    }
    /**
     * Returns the child node that has the label `variant`.
     */
    get variant() {
        return this.fetch();
    }
}
/**
 * This node represents a `StringExpression` nonterminal, with the following structure:
 *
 * ```ebnf
 * StringExpression = (* variant: *) StringLiteral (* Deprecated in 0.5.14 *)
 *                  | (* variant: *) StringLiterals (* Introduced in 0.5.14 *)
 *                  | (* variant: *) HexStringLiteral (* Deprecated in 0.5.14 *)
 *                  | (* variant: *) HexStringLiterals (* Introduced in 0.5.14 *)
 *                  | (* variant: *) UnicodeStringLiterals; (* Introduced in 0.7.0 *)
 * ```
 */
export class StringExpression {
    /**
     * Constructs a new AST node of type `StringExpression`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `StringExpression`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const variant = wasm.ast.Selectors.choice(this.cst);
            switch (variant.kind) {
                case NonterminalKind.StringLiteral:
                    return new StringLiteral(variant);
                case NonterminalKind.StringLiterals:
                    return new StringLiterals(variant);
                case NonterminalKind.HexStringLiteral:
                    return new HexStringLiteral(variant);
                case NonterminalKind.HexStringLiterals:
                    return new HexStringLiterals(variant);
                case NonterminalKind.UnicodeStringLiterals:
                    return new UnicodeStringLiterals(variant);
                default:
                    throw new Error(`Unexpected variant: '${variant.kind}'.`);
            }
        });
        assertKind(this.cst.kind, NonterminalKind.StringExpression);
    }
    /**
     * Returns the child node that has the label `variant`.
     */
    get variant() {
        return this.fetch();
    }
}
/**
 * This node represents a `StringLiteral` nonterminal, with the following structure:
 *
 * ```ebnf
 * StringLiteral = (* variant: *) SINGLE_QUOTED_STRING_LITERAL
 *               | (* variant: *) DOUBLE_QUOTED_STRING_LITERAL;
 * ```
 */
export class StringLiteral {
    /**
     * Constructs a new AST node of type `StringLiteral`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `StringLiteral`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const variant = wasm.ast.Selectors.choice(this.cst);
            return variant;
        });
        assertKind(this.cst.kind, NonterminalKind.StringLiteral);
    }
    /**
     * Returns the child node that has the label `variant`.
     */
    get variant() {
        return this.fetch();
    }
}
/**
 * This node represents a `HexStringLiteral` nonterminal, with the following structure:
 *
 * ```ebnf
 * HexStringLiteral = (* variant: *) SINGLE_QUOTED_HEX_STRING_LITERAL
 *                  | (* variant: *) DOUBLE_QUOTED_HEX_STRING_LITERAL;
 * ```
 */
export class HexStringLiteral {
    /**
     * Constructs a new AST node of type `HexStringLiteral`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `HexStringLiteral`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const variant = wasm.ast.Selectors.choice(this.cst);
            return variant;
        });
        assertKind(this.cst.kind, NonterminalKind.HexStringLiteral);
    }
    /**
     * Returns the child node that has the label `variant`.
     */
    get variant() {
        return this.fetch();
    }
}
/**
 * This node represents a `UnicodeStringLiteral` nonterminal, with the following structure:
 *
 * ```ebnf
 * (* Introduced in 0.7.0 *)
 * UnicodeStringLiteral = (* variant: *) SINGLE_QUOTED_UNICODE_STRING_LITERAL
 *                      | (* variant: *) DOUBLE_QUOTED_UNICODE_STRING_LITERAL;
 * ```
 */
export class UnicodeStringLiteral {
    /**
     * Constructs a new AST node of type `UnicodeStringLiteral`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `UnicodeStringLiteral`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const variant = wasm.ast.Selectors.choice(this.cst);
            return variant;
        });
        assertKind(this.cst.kind, NonterminalKind.UnicodeStringLiteral);
    }
    /**
     * Returns the child node that has the label `variant`.
     */
    get variant() {
        return this.fetch();
    }
}
/**
 * This node represents a `YulStatement` nonterminal, with the following structure:
 *
 * ```ebnf
 * YulStatement = (* variant: *) YulBlock
 *              | (* variant: *) YulFunctionDefinition
 *              | (* variant: *) YulStackAssignmentStatement (* Deprecated in 0.5.0 *)
 *              | (* variant: *) YulIfStatement
 *              | (* variant: *) YulForStatement
 *              | (* variant: *) YulSwitchStatement
 *              | (* variant: *) YulLeaveStatement (* Introduced in 0.6.0 *)
 *              | (* variant: *) YulBreakStatement
 *              | (* variant: *) YulContinueStatement
 *              | (* variant: *) YulVariableAssignmentStatement
 *              | (* variant: *) YulLabel (* Deprecated in 0.5.0 *)
 *              | (* variant: *) YulVariableDeclarationStatement
 *              | (* variant: *) YulExpression;
 * ```
 */
export class YulStatement {
    /**
     * Constructs a new AST node of type `YulStatement`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `YulStatement`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const variant = wasm.ast.Selectors.choice(this.cst);
            switch (variant.kind) {
                case NonterminalKind.YulBlock:
                    return new YulBlock(variant);
                case NonterminalKind.YulFunctionDefinition:
                    return new YulFunctionDefinition(variant);
                case NonterminalKind.YulStackAssignmentStatement:
                    return new YulStackAssignmentStatement(variant);
                case NonterminalKind.YulIfStatement:
                    return new YulIfStatement(variant);
                case NonterminalKind.YulForStatement:
                    return new YulForStatement(variant);
                case NonterminalKind.YulSwitchStatement:
                    return new YulSwitchStatement(variant);
                case NonterminalKind.YulLeaveStatement:
                    return new YulLeaveStatement(variant);
                case NonterminalKind.YulBreakStatement:
                    return new YulBreakStatement(variant);
                case NonterminalKind.YulContinueStatement:
                    return new YulContinueStatement(variant);
                case NonterminalKind.YulVariableAssignmentStatement:
                    return new YulVariableAssignmentStatement(variant);
                case NonterminalKind.YulLabel:
                    return new YulLabel(variant);
                case NonterminalKind.YulVariableDeclarationStatement:
                    return new YulVariableDeclarationStatement(variant);
                case NonterminalKind.YulExpression:
                    return new YulExpression(variant);
                default:
                    throw new Error(`Unexpected variant: '${variant.kind}'.`);
            }
        });
        assertKind(this.cst.kind, NonterminalKind.YulStatement);
    }
    /**
     * Returns the child node that has the label `variant`.
     */
    get variant() {
        return this.fetch();
    }
}
/**
 * This node represents a `YulAssignmentOperator` nonterminal, with the following structure:
 *
 * ```ebnf
 * YulAssignmentOperator = (* variant: *) COLON_EQUAL
 *                       | (* variant: *) YulColonAndEqual; (* Deprecated in 0.5.5 *)
 * ```
 */
export class YulAssignmentOperator {
    /**
     * Constructs a new AST node of type `YulAssignmentOperator`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `YulAssignmentOperator`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const variant = wasm.ast.Selectors.choice(this.cst);
            if (variant.isTerminalNode()) {
                return variant;
            }
            switch (variant.kind) {
                case NonterminalKind.YulColonAndEqual:
                    return new YulColonAndEqual(variant);
                default:
                    throw new Error(`Unexpected variant: '${variant.kind}'.`);
            }
        });
        assertKind(this.cst.kind, NonterminalKind.YulAssignmentOperator);
    }
    /**
     * Returns the child node that has the label `variant`.
     */
    get variant() {
        return this.fetch();
    }
}
/**
 * This node represents a `YulStackAssignmentOperator` nonterminal, with the following structure:
 *
 * ```ebnf
 * (* Deprecated in 0.5.0 *)
 * YulStackAssignmentOperator = (* variant: *) EQUAL_COLON
 *                            | (* variant: *) YulEqualAndColon;
 * ```
 */
export class YulStackAssignmentOperator {
    /**
     * Constructs a new AST node of type `YulStackAssignmentOperator`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `YulStackAssignmentOperator`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const variant = wasm.ast.Selectors.choice(this.cst);
            if (variant.isTerminalNode()) {
                return variant;
            }
            switch (variant.kind) {
                case NonterminalKind.YulEqualAndColon:
                    return new YulEqualAndColon(variant);
                default:
                    throw new Error(`Unexpected variant: '${variant.kind}'.`);
            }
        });
        assertKind(this.cst.kind, NonterminalKind.YulStackAssignmentOperator);
    }
    /**
     * Returns the child node that has the label `variant`.
     */
    get variant() {
        return this.fetch();
    }
}
/**
 * This node represents a `YulSwitchCase` nonterminal, with the following structure:
 *
 * ```ebnf
 * YulSwitchCase = (* variant: *) YulDefaultCase
 *               | (* variant: *) YulValueCase;
 * ```
 */
export class YulSwitchCase {
    /**
     * Constructs a new AST node of type `YulSwitchCase`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `YulSwitchCase`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const variant = wasm.ast.Selectors.choice(this.cst);
            switch (variant.kind) {
                case NonterminalKind.YulDefaultCase:
                    return new YulDefaultCase(variant);
                case NonterminalKind.YulValueCase:
                    return new YulValueCase(variant);
                default:
                    throw new Error(`Unexpected variant: '${variant.kind}'.`);
            }
        });
        assertKind(this.cst.kind, NonterminalKind.YulSwitchCase);
    }
    /**
     * Returns the child node that has the label `variant`.
     */
    get variant() {
        return this.fetch();
    }
}
/**
 * This node represents a `YulExpression` nonterminal, with the following structure:
 *
 * ```ebnf
 * YulExpression = (* variant: *) YulFunctionCallExpression
 *               | (* variant: *) YulLiteral
 *               | (* variant: *) YulPath;
 * ```
 */
export class YulExpression {
    /**
     * Constructs a new AST node of type `YulExpression`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `YulExpression`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const variant = wasm.ast.Selectors.choice(this.cst);
            switch (variant.kind) {
                case NonterminalKind.YulFunctionCallExpression:
                    return new YulFunctionCallExpression(variant);
                case NonterminalKind.YulLiteral:
                    return new YulLiteral(variant);
                case NonterminalKind.YulPath:
                    return new YulPath(variant);
                default:
                    throw new Error(`Unexpected variant: '${variant.kind}'.`);
            }
        });
        assertKind(this.cst.kind, NonterminalKind.YulExpression);
    }
    /**
     * Returns the child node that has the label `variant`.
     */
    get variant() {
        return this.fetch();
    }
}
/**
 * This node represents a `YulLiteral` nonterminal, with the following structure:
 *
 * ```ebnf
 * YulLiteral = (* variant: *) YUL_TRUE_KEYWORD (* Introduced in 0.6.2 *)
 *            | (* variant: *) YUL_FALSE_KEYWORD (* Introduced in 0.6.2 *)
 *            | (* variant: *) YUL_DECIMAL_LITERAL
 *            | (* variant: *) YUL_HEX_LITERAL
 *            | (* variant: *) HexStringLiteral
 *            | (* variant: *) StringLiteral;
 * ```
 */
export class YulLiteral {
    /**
     * Constructs a new AST node of type `YulLiteral`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `YulLiteral`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const variant = wasm.ast.Selectors.choice(this.cst);
            if (variant.isTerminalNode()) {
                return variant;
            }
            switch (variant.kind) {
                case NonterminalKind.HexStringLiteral:
                    return new HexStringLiteral(variant);
                case NonterminalKind.StringLiteral:
                    return new StringLiteral(variant);
                default:
                    throw new Error(`Unexpected variant: '${variant.kind}'.`);
            }
        });
        assertKind(this.cst.kind, NonterminalKind.YulLiteral);
    }
    /**
     * Returns the child node that has the label `variant`.
     */
    get variant() {
        return this.fetch();
    }
}
//
// Repeated:
//
/**
 * This node represents a `SourceUnitMembers` nonterminal, with the following structure:
 *
 * ```ebnf
 * SourceUnitMembers = (* item: *) SourceUnitMember*;
 * ```
 */
export class SourceUnitMembers {
    /**
     * Constructs a new AST node of type `SourceUnitMembers`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `SourceUnitMembers`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const items = wasm.ast.Selectors.repeated(this.cst);
            return items.map((item) => new SourceUnitMember(item));
        });
        assertKind(this.cst.kind, NonterminalKind.SourceUnitMembers);
    }
    /**
     * Returns an array of the child nodes that have the label `item`.
     */
    get items() {
        return this.fetch();
    }
}
/**
 * This node represents a `VersionExpressionSet` nonterminal, with the following structure:
 *
 * ```ebnf
 * VersionExpressionSet = (* item: *) VersionExpression+;
 * ```
 */
export class VersionExpressionSet {
    /**
     * Constructs a new AST node of type `VersionExpressionSet`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `VersionExpressionSet`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const items = wasm.ast.Selectors.repeated(this.cst);
            return items.map((item) => new VersionExpression(item));
        });
        assertKind(this.cst.kind, NonterminalKind.VersionExpressionSet);
    }
    /**
     * Returns an array of the child nodes that have the label `item`.
     */
    get items() {
        return this.fetch();
    }
}
/**
 * This node represents a `ContractSpecifiers` nonterminal, with the following structure:
 *
 * ```ebnf
 * ContractSpecifiers = (* item: *) ContractSpecifier*;
 * ```
 */
export class ContractSpecifiers {
    /**
     * Constructs a new AST node of type `ContractSpecifiers`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `ContractSpecifiers`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const items = wasm.ast.Selectors.repeated(this.cst);
            return items.map((item) => new ContractSpecifier(item));
        });
        assertKind(this.cst.kind, NonterminalKind.ContractSpecifiers);
    }
    /**
     * Returns an array of the child nodes that have the label `item`.
     */
    get items() {
        return this.fetch();
    }
}
/**
 * This node represents a `ContractMembers` nonterminal, with the following structure:
 *
 * ```ebnf
 * ContractMembers = (* item: *) ContractMember*;
 * ```
 */
export class ContractMembers {
    /**
     * Constructs a new AST node of type `ContractMembers`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `ContractMembers`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const items = wasm.ast.Selectors.repeated(this.cst);
            return items.map((item) => new ContractMember(item));
        });
        assertKind(this.cst.kind, NonterminalKind.ContractMembers);
    }
    /**
     * Returns an array of the child nodes that have the label `item`.
     */
    get items() {
        return this.fetch();
    }
}
/**
 * This node represents a `InterfaceMembers` nonterminal, with the following structure:
 *
 * ```ebnf
 * InterfaceMembers = (* item: *) ContractMember*;
 * ```
 */
export class InterfaceMembers {
    /**
     * Constructs a new AST node of type `InterfaceMembers`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `InterfaceMembers`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const items = wasm.ast.Selectors.repeated(this.cst);
            return items.map((item) => new ContractMember(item));
        });
        assertKind(this.cst.kind, NonterminalKind.InterfaceMembers);
    }
    /**
     * Returns an array of the child nodes that have the label `item`.
     */
    get items() {
        return this.fetch();
    }
}
/**
 * This node represents a `LibraryMembers` nonterminal, with the following structure:
 *
 * ```ebnf
 * LibraryMembers = (* item: *) ContractMember*;
 * ```
 */
export class LibraryMembers {
    /**
     * Constructs a new AST node of type `LibraryMembers`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `LibraryMembers`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const items = wasm.ast.Selectors.repeated(this.cst);
            return items.map((item) => new ContractMember(item));
        });
        assertKind(this.cst.kind, NonterminalKind.LibraryMembers);
    }
    /**
     * Returns an array of the child nodes that have the label `item`.
     */
    get items() {
        return this.fetch();
    }
}
/**
 * This node represents a `StructMembers` nonterminal, with the following structure:
 *
 * ```ebnf
 * StructMembers = (* item: *) StructMember*;
 * ```
 */
export class StructMembers {
    /**
     * Constructs a new AST node of type `StructMembers`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `StructMembers`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const items = wasm.ast.Selectors.repeated(this.cst);
            return items.map((item) => new StructMember(item));
        });
        assertKind(this.cst.kind, NonterminalKind.StructMembers);
    }
    /**
     * Returns an array of the child nodes that have the label `item`.
     */
    get items() {
        return this.fetch();
    }
}
/**
 * This node represents a `StateVariableAttributes` nonterminal, with the following structure:
 *
 * ```ebnf
 * StateVariableAttributes = (* item: *) StateVariableAttribute*;
 * ```
 */
export class StateVariableAttributes {
    /**
     * Constructs a new AST node of type `StateVariableAttributes`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `StateVariableAttributes`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const items = wasm.ast.Selectors.repeated(this.cst);
            return items.map((item) => new StateVariableAttribute(item));
        });
        assertKind(this.cst.kind, NonterminalKind.StateVariableAttributes);
    }
    /**
     * Returns an array of the child nodes that have the label `item`.
     */
    get items() {
        return this.fetch();
    }
}
/**
 * This node represents a `FunctionAttributes` nonterminal, with the following structure:
 *
 * ```ebnf
 * FunctionAttributes = (* item: *) FunctionAttribute*;
 * ```
 */
export class FunctionAttributes {
    /**
     * Constructs a new AST node of type `FunctionAttributes`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `FunctionAttributes`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const items = wasm.ast.Selectors.repeated(this.cst);
            return items.map((item) => new FunctionAttribute(item));
        });
        assertKind(this.cst.kind, NonterminalKind.FunctionAttributes);
    }
    /**
     * Returns an array of the child nodes that have the label `item`.
     */
    get items() {
        return this.fetch();
    }
}
/**
 * This node represents a `ConstructorAttributes` nonterminal, with the following structure:
 *
 * ```ebnf
 * (* Introduced in 0.4.22 *)
 * ConstructorAttributes = (* item: *) ConstructorAttribute*;
 * ```
 */
export class ConstructorAttributes {
    /**
     * Constructs a new AST node of type `ConstructorAttributes`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `ConstructorAttributes`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const items = wasm.ast.Selectors.repeated(this.cst);
            return items.map((item) => new ConstructorAttribute(item));
        });
        assertKind(this.cst.kind, NonterminalKind.ConstructorAttributes);
    }
    /**
     * Returns an array of the child nodes that have the label `item`.
     */
    get items() {
        return this.fetch();
    }
}
/**
 * This node represents a `UnnamedFunctionAttributes` nonterminal, with the following structure:
 *
 * ```ebnf
 * (* Deprecated in 0.6.0 *)
 * UnnamedFunctionAttributes = (* item: *) UnnamedFunctionAttribute*;
 * ```
 */
export class UnnamedFunctionAttributes {
    /**
     * Constructs a new AST node of type `UnnamedFunctionAttributes`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `UnnamedFunctionAttributes`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const items = wasm.ast.Selectors.repeated(this.cst);
            return items.map((item) => new UnnamedFunctionAttribute(item));
        });
        assertKind(this.cst.kind, NonterminalKind.UnnamedFunctionAttributes);
    }
    /**
     * Returns an array of the child nodes that have the label `item`.
     */
    get items() {
        return this.fetch();
    }
}
/**
 * This node represents a `FallbackFunctionAttributes` nonterminal, with the following structure:
 *
 * ```ebnf
 * (* Introduced in 0.6.0 *)
 * FallbackFunctionAttributes = (* item: *) FallbackFunctionAttribute*;
 * ```
 */
export class FallbackFunctionAttributes {
    /**
     * Constructs a new AST node of type `FallbackFunctionAttributes`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `FallbackFunctionAttributes`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const items = wasm.ast.Selectors.repeated(this.cst);
            return items.map((item) => new FallbackFunctionAttribute(item));
        });
        assertKind(this.cst.kind, NonterminalKind.FallbackFunctionAttributes);
    }
    /**
     * Returns an array of the child nodes that have the label `item`.
     */
    get items() {
        return this.fetch();
    }
}
/**
 * This node represents a `ReceiveFunctionAttributes` nonterminal, with the following structure:
 *
 * ```ebnf
 * (* Introduced in 0.6.0 *)
 * ReceiveFunctionAttributes = (* item: *) ReceiveFunctionAttribute*;
 * ```
 */
export class ReceiveFunctionAttributes {
    /**
     * Constructs a new AST node of type `ReceiveFunctionAttributes`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `ReceiveFunctionAttributes`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const items = wasm.ast.Selectors.repeated(this.cst);
            return items.map((item) => new ReceiveFunctionAttribute(item));
        });
        assertKind(this.cst.kind, NonterminalKind.ReceiveFunctionAttributes);
    }
    /**
     * Returns an array of the child nodes that have the label `item`.
     */
    get items() {
        return this.fetch();
    }
}
/**
 * This node represents a `ModifierAttributes` nonterminal, with the following structure:
 *
 * ```ebnf
 * ModifierAttributes = (* item: *) ModifierAttribute*;
 * ```
 */
export class ModifierAttributes {
    /**
     * Constructs a new AST node of type `ModifierAttributes`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `ModifierAttributes`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const items = wasm.ast.Selectors.repeated(this.cst);
            return items.map((item) => new ModifierAttribute(item));
        });
        assertKind(this.cst.kind, NonterminalKind.ModifierAttributes);
    }
    /**
     * Returns an array of the child nodes that have the label `item`.
     */
    get items() {
        return this.fetch();
    }
}
/**
 * This node represents a `FunctionTypeAttributes` nonterminal, with the following structure:
 *
 * ```ebnf
 * FunctionTypeAttributes = (* item: *) FunctionTypeAttribute*;
 * ```
 */
export class FunctionTypeAttributes {
    /**
     * Constructs a new AST node of type `FunctionTypeAttributes`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `FunctionTypeAttributes`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const items = wasm.ast.Selectors.repeated(this.cst);
            return items.map((item) => new FunctionTypeAttribute(item));
        });
        assertKind(this.cst.kind, NonterminalKind.FunctionTypeAttributes);
    }
    /**
     * Returns an array of the child nodes that have the label `item`.
     */
    get items() {
        return this.fetch();
    }
}
/**
 * This node represents a `Statements` nonterminal, with the following structure:
 *
 * ```ebnf
 * Statements = (* item: *) Statement*;
 * ```
 */
export class Statements {
    /**
     * Constructs a new AST node of type `Statements`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `Statements`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const items = wasm.ast.Selectors.repeated(this.cst);
            return items.map((item) => new Statement(item));
        });
        assertKind(this.cst.kind, NonterminalKind.Statements);
    }
    /**
     * Returns an array of the child nodes that have the label `item`.
     */
    get items() {
        return this.fetch();
    }
}
/**
 * This node represents a `CatchClauses` nonterminal, with the following structure:
 *
 * ```ebnf
 * (* Introduced in 0.6.0 *)
 * CatchClauses = (* item: *) CatchClause+;
 * ```
 */
export class CatchClauses {
    /**
     * Constructs a new AST node of type `CatchClauses`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `CatchClauses`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const items = wasm.ast.Selectors.repeated(this.cst);
            return items.map((item) => new CatchClause(item));
        });
        assertKind(this.cst.kind, NonterminalKind.CatchClauses);
    }
    /**
     * Returns an array of the child nodes that have the label `item`.
     */
    get items() {
        return this.fetch();
    }
}
/**
 * This node represents a `StringLiterals` nonterminal, with the following structure:
 *
 * ```ebnf
 * (* Introduced in 0.5.14 *)
 * StringLiterals = (* item: *) StringLiteral+;
 * ```
 */
export class StringLiterals {
    /**
     * Constructs a new AST node of type `StringLiterals`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `StringLiterals`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const items = wasm.ast.Selectors.repeated(this.cst);
            return items.map((item) => new StringLiteral(item));
        });
        assertKind(this.cst.kind, NonterminalKind.StringLiterals);
    }
    /**
     * Returns an array of the child nodes that have the label `item`.
     */
    get items() {
        return this.fetch();
    }
}
/**
 * This node represents a `HexStringLiterals` nonterminal, with the following structure:
 *
 * ```ebnf
 * (* Introduced in 0.5.14 *)
 * HexStringLiterals = (* item: *) HexStringLiteral+;
 * ```
 */
export class HexStringLiterals {
    /**
     * Constructs a new AST node of type `HexStringLiterals`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `HexStringLiterals`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const items = wasm.ast.Selectors.repeated(this.cst);
            return items.map((item) => new HexStringLiteral(item));
        });
        assertKind(this.cst.kind, NonterminalKind.HexStringLiterals);
    }
    /**
     * Returns an array of the child nodes that have the label `item`.
     */
    get items() {
        return this.fetch();
    }
}
/**
 * This node represents a `UnicodeStringLiterals` nonterminal, with the following structure:
 *
 * ```ebnf
 * (* Introduced in 0.7.0 *)
 * UnicodeStringLiterals = (* item: *) UnicodeStringLiteral+;
 * ```
 */
export class UnicodeStringLiterals {
    /**
     * Constructs a new AST node of type `UnicodeStringLiterals`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `UnicodeStringLiterals`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const items = wasm.ast.Selectors.repeated(this.cst);
            return items.map((item) => new UnicodeStringLiteral(item));
        });
        assertKind(this.cst.kind, NonterminalKind.UnicodeStringLiterals);
    }
    /**
     * Returns an array of the child nodes that have the label `item`.
     */
    get items() {
        return this.fetch();
    }
}
/**
 * This node represents a `YulStatements` nonterminal, with the following structure:
 *
 * ```ebnf
 * YulStatements = (* item: *) YulStatement*;
 * ```
 */
export class YulStatements {
    /**
     * Constructs a new AST node of type `YulStatements`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `YulStatements`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const items = wasm.ast.Selectors.repeated(this.cst);
            return items.map((item) => new YulStatement(item));
        });
        assertKind(this.cst.kind, NonterminalKind.YulStatements);
    }
    /**
     * Returns an array of the child nodes that have the label `item`.
     */
    get items() {
        return this.fetch();
    }
}
/**
 * This node represents a `YulSwitchCases` nonterminal, with the following structure:
 *
 * ```ebnf
 * YulSwitchCases = (* item: *) YulSwitchCase+;
 * ```
 */
export class YulSwitchCases {
    /**
     * Constructs a new AST node of type `YulSwitchCases`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `YulSwitchCases`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const items = wasm.ast.Selectors.repeated(this.cst);
            return items.map((item) => new YulSwitchCase(item));
        });
        assertKind(this.cst.kind, NonterminalKind.YulSwitchCases);
    }
    /**
     * Returns an array of the child nodes that have the label `item`.
     */
    get items() {
        return this.fetch();
    }
}
//
// Separated:
//
/**
 * This node represents a `VersionExpressionSets` nonterminal, with the following structure:
 *
 * ```ebnf
 * VersionExpressionSets = (* item: *) VersionExpressionSet ((* separator: *) BAR_BAR (* item: *) VersionExpressionSet)*;
 * ```
 */
export class VersionExpressionSets {
    /**
     * Constructs a new AST node of type `VersionExpressionSets`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `VersionExpressionSets`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const [items, separators] = wasm.ast.Selectors.separated(this.cst);
            return {
                items: items.map((item) => new VersionExpressionSet(item)),
                separators: separators,
            };
        });
        assertKind(this.cst.kind, NonterminalKind.VersionExpressionSets);
    }
    /**
     * Returns an array of the child nodes that have the label `item`.
     */
    get items() {
        return this.fetch().items;
    }
    /**
     * Returns an array of the child nodes that have the label `separator`.
     */
    get separators() {
        return this.fetch().separators;
    }
}
/**
 * This node represents a `SimpleVersionLiteral` nonterminal, with the following structure:
 *
 * ```ebnf
 * SimpleVersionLiteral = (* item: *) VERSION_SPECIFIER ((* separator: *) PERIOD (* item: *) VERSION_SPECIFIER)*;
 * ```
 */
export class SimpleVersionLiteral {
    /**
     * Constructs a new AST node of type `SimpleVersionLiteral`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `SimpleVersionLiteral`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const [items, separators] = wasm.ast.Selectors.separated(this.cst);
            return { items: items, separators: separators };
        });
        assertKind(this.cst.kind, NonterminalKind.SimpleVersionLiteral);
    }
    /**
     * Returns an array of the child nodes that have the label `item`.
     */
    get items() {
        return this.fetch().items;
    }
    /**
     * Returns an array of the child nodes that have the label `separator`.
     */
    get separators() {
        return this.fetch().separators;
    }
}
/**
 * This node represents a `ImportDeconstructionSymbols` nonterminal, with the following structure:
 *
 * ```ebnf
 * ImportDeconstructionSymbols = (* item: *) ImportDeconstructionSymbol ((* separator: *) COMMA (* item: *) ImportDeconstructionSymbol)*;
 * ```
 */
export class ImportDeconstructionSymbols {
    /**
     * Constructs a new AST node of type `ImportDeconstructionSymbols`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `ImportDeconstructionSymbols`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const [items, separators] = wasm.ast.Selectors.separated(this.cst);
            return {
                items: items.map((item) => new ImportDeconstructionSymbol(item)),
                separators: separators,
            };
        });
        assertKind(this.cst.kind, NonterminalKind.ImportDeconstructionSymbols);
    }
    /**
     * Returns an array of the child nodes that have the label `item`.
     */
    get items() {
        return this.fetch().items;
    }
    /**
     * Returns an array of the child nodes that have the label `separator`.
     */
    get separators() {
        return this.fetch().separators;
    }
}
/**
 * This node represents a `UsingDeconstructionSymbols` nonterminal, with the following structure:
 *
 * ```ebnf
 * (* Introduced in 0.8.13 *)
 * UsingDeconstructionSymbols = (* item: *) UsingDeconstructionSymbol ((* separator: *) COMMA (* item: *) UsingDeconstructionSymbol)*;
 * ```
 */
export class UsingDeconstructionSymbols {
    /**
     * Constructs a new AST node of type `UsingDeconstructionSymbols`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `UsingDeconstructionSymbols`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const [items, separators] = wasm.ast.Selectors.separated(this.cst);
            return {
                items: items.map((item) => new UsingDeconstructionSymbol(item)),
                separators: separators,
            };
        });
        assertKind(this.cst.kind, NonterminalKind.UsingDeconstructionSymbols);
    }
    /**
     * Returns an array of the child nodes that have the label `item`.
     */
    get items() {
        return this.fetch().items;
    }
    /**
     * Returns an array of the child nodes that have the label `separator`.
     */
    get separators() {
        return this.fetch().separators;
    }
}
/**
 * This node represents a `InheritanceTypes` nonterminal, with the following structure:
 *
 * ```ebnf
 * InheritanceTypes = (* item: *) InheritanceType ((* separator: *) COMMA (* item: *) InheritanceType)*;
 * ```
 */
export class InheritanceTypes {
    /**
     * Constructs a new AST node of type `InheritanceTypes`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `InheritanceTypes`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const [items, separators] = wasm.ast.Selectors.separated(this.cst);
            return {
                items: items.map((item) => new InheritanceType(item)),
                separators: separators,
            };
        });
        assertKind(this.cst.kind, NonterminalKind.InheritanceTypes);
    }
    /**
     * Returns an array of the child nodes that have the label `item`.
     */
    get items() {
        return this.fetch().items;
    }
    /**
     * Returns an array of the child nodes that have the label `separator`.
     */
    get separators() {
        return this.fetch().separators;
    }
}
/**
 * This node represents a `EnumMembers` nonterminal, with the following structure:
 *
 * ```ebnf
 * EnumMembers = ((* item: *) IDENTIFIER ((* separator: *) COMMA (* item: *) IDENTIFIER)*)?;
 * ```
 */
export class EnumMembers {
    /**
     * Constructs a new AST node of type `EnumMembers`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `EnumMembers`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const [items, separators] = wasm.ast.Selectors.separated(this.cst);
            return { items: items, separators: separators };
        });
        assertKind(this.cst.kind, NonterminalKind.EnumMembers);
    }
    /**
     * Returns an array of the child nodes that have the label `item`.
     */
    get items() {
        return this.fetch().items;
    }
    /**
     * Returns an array of the child nodes that have the label `separator`.
     */
    get separators() {
        return this.fetch().separators;
    }
}
/**
 * This node represents a `Parameters` nonterminal, with the following structure:
 *
 * ```ebnf
 * Parameters = ((* item: *) Parameter ((* separator: *) COMMA (* item: *) Parameter)*)?;
 * ```
 */
export class Parameters {
    /**
     * Constructs a new AST node of type `Parameters`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `Parameters`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const [items, separators] = wasm.ast.Selectors.separated(this.cst);
            return {
                items: items.map((item) => new Parameter(item)),
                separators: separators,
            };
        });
        assertKind(this.cst.kind, NonterminalKind.Parameters);
    }
    /**
     * Returns an array of the child nodes that have the label `item`.
     */
    get items() {
        return this.fetch().items;
    }
    /**
     * Returns an array of the child nodes that have the label `separator`.
     */
    get separators() {
        return this.fetch().separators;
    }
}
/**
 * This node represents a `OverridePaths` nonterminal, with the following structure:
 *
 * ```ebnf
 * (* Introduced in 0.6.0 *)
 * OverridePaths = (* item: *) IdentifierPath ((* separator: *) COMMA (* item: *) IdentifierPath)*;
 * ```
 */
export class OverridePaths {
    /**
     * Constructs a new AST node of type `OverridePaths`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `OverridePaths`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const [items, separators] = wasm.ast.Selectors.separated(this.cst);
            return {
                items: items.map((item) => new IdentifierPath(item)),
                separators: separators,
            };
        });
        assertKind(this.cst.kind, NonterminalKind.OverridePaths);
    }
    /**
     * Returns an array of the child nodes that have the label `item`.
     */
    get items() {
        return this.fetch().items;
    }
    /**
     * Returns an array of the child nodes that have the label `separator`.
     */
    get separators() {
        return this.fetch().separators;
    }
}
/**
 * This node represents a `EventParameters` nonterminal, with the following structure:
 *
 * ```ebnf
 * EventParameters = ((* item: *) EventParameter ((* separator: *) COMMA (* item: *) EventParameter)*)?;
 * ```
 */
export class EventParameters {
    /**
     * Constructs a new AST node of type `EventParameters`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `EventParameters`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const [items, separators] = wasm.ast.Selectors.separated(this.cst);
            return {
                items: items.map((item) => new EventParameter(item)),
                separators: separators,
            };
        });
        assertKind(this.cst.kind, NonterminalKind.EventParameters);
    }
    /**
     * Returns an array of the child nodes that have the label `item`.
     */
    get items() {
        return this.fetch().items;
    }
    /**
     * Returns an array of the child nodes that have the label `separator`.
     */
    get separators() {
        return this.fetch().separators;
    }
}
/**
 * This node represents a `ErrorParameters` nonterminal, with the following structure:
 *
 * ```ebnf
 * (* Introduced in 0.8.4 *)
 * ErrorParameters = ((* item: *) ErrorParameter ((* separator: *) COMMA (* item: *) ErrorParameter)*)?;
 * ```
 */
export class ErrorParameters {
    /**
     * Constructs a new AST node of type `ErrorParameters`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `ErrorParameters`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const [items, separators] = wasm.ast.Selectors.separated(this.cst);
            return {
                items: items.map((item) => new ErrorParameter(item)),
                separators: separators,
            };
        });
        assertKind(this.cst.kind, NonterminalKind.ErrorParameters);
    }
    /**
     * Returns an array of the child nodes that have the label `item`.
     */
    get items() {
        return this.fetch().items;
    }
    /**
     * Returns an array of the child nodes that have the label `separator`.
     */
    get separators() {
        return this.fetch().separators;
    }
}
/**
 * This node represents a `AssemblyFlags` nonterminal, with the following structure:
 *
 * ```ebnf
 * (* Introduced in 0.8.13 *)
 * AssemblyFlags = (* item: *) StringLiteral ((* separator: *) COMMA (* item: *) StringLiteral)*;
 * ```
 */
export class AssemblyFlags {
    /**
     * Constructs a new AST node of type `AssemblyFlags`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `AssemblyFlags`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const [items, separators] = wasm.ast.Selectors.separated(this.cst);
            return {
                items: items.map((item) => new StringLiteral(item)),
                separators: separators,
            };
        });
        assertKind(this.cst.kind, NonterminalKind.AssemblyFlags);
    }
    /**
     * Returns an array of the child nodes that have the label `item`.
     */
    get items() {
        return this.fetch().items;
    }
    /**
     * Returns an array of the child nodes that have the label `separator`.
     */
    get separators() {
        return this.fetch().separators;
    }
}
/**
 * This node represents a `TupleDeconstructionElements` nonterminal, with the following structure:
 *
 * ```ebnf
 * TupleDeconstructionElements = (* item: *) TupleDeconstructionElement ((* separator: *) COMMA (* item: *) TupleDeconstructionElement)*;
 * ```
 */
export class TupleDeconstructionElements {
    /**
     * Constructs a new AST node of type `TupleDeconstructionElements`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `TupleDeconstructionElements`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const [items, separators] = wasm.ast.Selectors.separated(this.cst);
            return {
                items: items.map((item) => new TupleDeconstructionElement(item)),
                separators: separators,
            };
        });
        assertKind(this.cst.kind, NonterminalKind.TupleDeconstructionElements);
    }
    /**
     * Returns an array of the child nodes that have the label `item`.
     */
    get items() {
        return this.fetch().items;
    }
    /**
     * Returns an array of the child nodes that have the label `separator`.
     */
    get separators() {
        return this.fetch().separators;
    }
}
/**
 * This node represents a `PositionalArguments` nonterminal, with the following structure:
 *
 * ```ebnf
 * PositionalArguments = ((* item: *) Expression ((* separator: *) COMMA (* item: *) Expression)*)?;
 * ```
 */
export class PositionalArguments {
    /**
     * Constructs a new AST node of type `PositionalArguments`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `PositionalArguments`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const [items, separators] = wasm.ast.Selectors.separated(this.cst);
            return {
                items: items.map((item) => new Expression(item)),
                separators: separators,
            };
        });
        assertKind(this.cst.kind, NonterminalKind.PositionalArguments);
    }
    /**
     * Returns an array of the child nodes that have the label `item`.
     */
    get items() {
        return this.fetch().items;
    }
    /**
     * Returns an array of the child nodes that have the label `separator`.
     */
    get separators() {
        return this.fetch().separators;
    }
}
/**
 * This node represents a `NamedArguments` nonterminal, with the following structure:
 *
 * ```ebnf
 * NamedArguments = ((* item: *) NamedArgument ((* separator: *) COMMA (* item: *) NamedArgument)*)?;
 * ```
 */
export class NamedArguments {
    /**
     * Constructs a new AST node of type `NamedArguments`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `NamedArguments`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const [items, separators] = wasm.ast.Selectors.separated(this.cst);
            return {
                items: items.map((item) => new NamedArgument(item)),
                separators: separators,
            };
        });
        assertKind(this.cst.kind, NonterminalKind.NamedArguments);
    }
    /**
     * Returns an array of the child nodes that have the label `item`.
     */
    get items() {
        return this.fetch().items;
    }
    /**
     * Returns an array of the child nodes that have the label `separator`.
     */
    get separators() {
        return this.fetch().separators;
    }
}
/**
 * This node represents a `CallOptions` nonterminal, with the following structure:
 *
 * ```ebnf
 * (* Introduced in 0.6.2 *)
 * CallOptions = (* item: *) NamedArgument ((* separator: *) COMMA (* item: *) NamedArgument)*;
 * ```
 */
export class CallOptions {
    /**
     * Constructs a new AST node of type `CallOptions`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `CallOptions`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const [items, separators] = wasm.ast.Selectors.separated(this.cst);
            return {
                items: items.map((item) => new NamedArgument(item)),
                separators: separators,
            };
        });
        assertKind(this.cst.kind, NonterminalKind.CallOptions);
    }
    /**
     * Returns an array of the child nodes that have the label `item`.
     */
    get items() {
        return this.fetch().items;
    }
    /**
     * Returns an array of the child nodes that have the label `separator`.
     */
    get separators() {
        return this.fetch().separators;
    }
}
/**
 * This node represents a `TupleValues` nonterminal, with the following structure:
 *
 * ```ebnf
 * TupleValues = (* item: *) TupleValue ((* separator: *) COMMA (* item: *) TupleValue)*;
 * ```
 */
export class TupleValues {
    /**
     * Constructs a new AST node of type `TupleValues`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `TupleValues`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const [items, separators] = wasm.ast.Selectors.separated(this.cst);
            return {
                items: items.map((item) => new TupleValue(item)),
                separators: separators,
            };
        });
        assertKind(this.cst.kind, NonterminalKind.TupleValues);
    }
    /**
     * Returns an array of the child nodes that have the label `item`.
     */
    get items() {
        return this.fetch().items;
    }
    /**
     * Returns an array of the child nodes that have the label `separator`.
     */
    get separators() {
        return this.fetch().separators;
    }
}
/**
 * This node represents a `ArrayValues` nonterminal, with the following structure:
 *
 * ```ebnf
 * ArrayValues = (* item: *) Expression ((* separator: *) COMMA (* item: *) Expression)*;
 * ```
 */
export class ArrayValues {
    /**
     * Constructs a new AST node of type `ArrayValues`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `ArrayValues`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const [items, separators] = wasm.ast.Selectors.separated(this.cst);
            return {
                items: items.map((item) => new Expression(item)),
                separators: separators,
            };
        });
        assertKind(this.cst.kind, NonterminalKind.ArrayValues);
    }
    /**
     * Returns an array of the child nodes that have the label `item`.
     */
    get items() {
        return this.fetch().items;
    }
    /**
     * Returns an array of the child nodes that have the label `separator`.
     */
    get separators() {
        return this.fetch().separators;
    }
}
/**
 * This node represents a `IdentifierPath` nonterminal, with the following structure:
 *
 * ```ebnf
 * IdentifierPath = (* item: *) IDENTIFIER ((* separator: *) PERIOD (* item: *) IDENTIFIER)*;
 * ```
 */
export class IdentifierPath {
    /**
     * Constructs a new AST node of type `IdentifierPath`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `IdentifierPath`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const [items, separators] = wasm.ast.Selectors.separated(this.cst);
            return { items: items, separators: separators };
        });
        assertKind(this.cst.kind, NonterminalKind.IdentifierPath);
    }
    /**
     * Returns an array of the child nodes that have the label `item`.
     */
    get items() {
        return this.fetch().items;
    }
    /**
     * Returns an array of the child nodes that have the label `separator`.
     */
    get separators() {
        return this.fetch().separators;
    }
}
/**
 * This node represents a `YulParameters` nonterminal, with the following structure:
 *
 * ```ebnf
 * YulParameters = ((* item: *) YUL_IDENTIFIER ((* separator: *) COMMA (* item: *) YUL_IDENTIFIER)*)?;
 * ```
 */
export class YulParameters {
    /**
     * Constructs a new AST node of type `YulParameters`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `YulParameters`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const [items, separators] = wasm.ast.Selectors.separated(this.cst);
            return { items: items, separators: separators };
        });
        assertKind(this.cst.kind, NonterminalKind.YulParameters);
    }
    /**
     * Returns an array of the child nodes that have the label `item`.
     */
    get items() {
        return this.fetch().items;
    }
    /**
     * Returns an array of the child nodes that have the label `separator`.
     */
    get separators() {
        return this.fetch().separators;
    }
}
/**
 * This node represents a `YulVariableNames` nonterminal, with the following structure:
 *
 * ```ebnf
 * YulVariableNames = (* item: *) YUL_IDENTIFIER ((* separator: *) COMMA (* item: *) YUL_IDENTIFIER)*;
 * ```
 */
export class YulVariableNames {
    /**
     * Constructs a new AST node of type `YulVariableNames`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `YulVariableNames`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const [items, separators] = wasm.ast.Selectors.separated(this.cst);
            return { items: items, separators: separators };
        });
        assertKind(this.cst.kind, NonterminalKind.YulVariableNames);
    }
    /**
     * Returns an array of the child nodes that have the label `item`.
     */
    get items() {
        return this.fetch().items;
    }
    /**
     * Returns an array of the child nodes that have the label `separator`.
     */
    get separators() {
        return this.fetch().separators;
    }
}
/**
 * This node represents a `YulArguments` nonterminal, with the following structure:
 *
 * ```ebnf
 * YulArguments = ((* item: *) YulExpression ((* separator: *) COMMA (* item: *) YulExpression)*)?;
 * ```
 */
export class YulArguments {
    /**
     * Constructs a new AST node of type `YulArguments`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `YulArguments`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const [items, separators] = wasm.ast.Selectors.separated(this.cst);
            return {
                items: items.map((item) => new YulExpression(item)),
                separators: separators,
            };
        });
        assertKind(this.cst.kind, NonterminalKind.YulArguments);
    }
    /**
     * Returns an array of the child nodes that have the label `item`.
     */
    get items() {
        return this.fetch().items;
    }
    /**
     * Returns an array of the child nodes that have the label `separator`.
     */
    get separators() {
        return this.fetch().separators;
    }
}
/**
 * This node represents a `YulPaths` nonterminal, with the following structure:
 *
 * ```ebnf
 * YulPaths = (* item: *) YulPath ((* separator: *) COMMA (* item: *) YulPath)*;
 * ```
 */
export class YulPaths {
    /**
     * Constructs a new AST node of type `YulPaths`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `YulPaths`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const [items, separators] = wasm.ast.Selectors.separated(this.cst);
            return {
                items: items.map((item) => new YulPath(item)),
                separators: separators,
            };
        });
        assertKind(this.cst.kind, NonterminalKind.YulPaths);
    }
    /**
     * Returns an array of the child nodes that have the label `item`.
     */
    get items() {
        return this.fetch().items;
    }
    /**
     * Returns an array of the child nodes that have the label `separator`.
     */
    get separators() {
        return this.fetch().separators;
    }
}
/**
 * This node represents a `YulPath` nonterminal, with the following structure:
 *
 * ```ebnf
 * YulPath = (* item: *) YUL_IDENTIFIER ((* separator: *) PERIOD (* item: *) YUL_IDENTIFIER)*;
 * ```
 */
export class YulPath {
    /**
     * Constructs a new AST node of type `YulPath`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `YulPath`.
     */
    cst) {
        this.cst = cst;
        this.fetch = once(() => {
            const [items, separators] = wasm.ast.Selectors.separated(this.cst);
            return { items: items, separators: separators };
        });
        assertKind(this.cst.kind, NonterminalKind.YulPath);
    }
    /**
     * Returns an array of the child nodes that have the label `item`.
     */
    get items() {
        return this.fetch().items;
    }
    /**
     * Returns an array of the child nodes that have the label `separator`.
     */
    get separators() {
        return this.fetch().separators;
    }
}
function once(factory) {
    let value;
    return () => {
        if (value === undefined) {
            value = factory();
        }
        return value;
    };
}
function assertKind(actual, expected) {
    if (actual !== expected) {
        throw new Error(`AST node '${expected}' can only be initialized with a CST node of the same kind. Received '${actual}' instead.`);
    }
}
//# sourceMappingURL=nodes.generated.mjs.map