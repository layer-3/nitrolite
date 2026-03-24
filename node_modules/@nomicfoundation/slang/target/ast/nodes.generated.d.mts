import { NonterminalNode, TerminalNode } from "../cst/index.mjs";
/**
 * This node represents a `SourceUnit` nonterminal, with the following structure:
 *
 * ```ebnf
 * SourceUnit = (* members: *) SourceUnitMembers;
 * ```
 */
export declare class SourceUnit {
    /**
     * The underlying nonterminal CST node of kind `SourceUnit`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `SourceUnit`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `SourceUnit`.
     */
    cst: NonterminalNode);
    /**
     * Returns the child node that has the label `members`.
     */
    get members(): SourceUnitMembers;
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
export declare class PragmaDirective {
    /**
     * The underlying nonterminal CST node of kind `PragmaDirective`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `PragmaDirective`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `PragmaDirective`.
     */
    cst: NonterminalNode);
    /**
     * Returns the child node that has the label `pragma_keyword`.
     */
    get pragmaKeyword(): TerminalNode;
    /**
     * Returns the child node that has the label `pragma`.
     */
    get pragma(): Pragma;
    /**
     * Returns the child node that has the label `semicolon`.
     */
    get semicolon(): TerminalNode;
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
export declare class AbicoderPragma {
    /**
     * The underlying nonterminal CST node of kind `AbicoderPragma`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `AbicoderPragma`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `AbicoderPragma`.
     */
    cst: NonterminalNode);
    /**
     * Returns the child node that has the label `abicoder_keyword`.
     */
    get abicoderKeyword(): TerminalNode;
    /**
     * Returns the child node that has the label `version`.
     */
    get version(): AbicoderVersion;
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
export declare class ExperimentalPragma {
    /**
     * The underlying nonterminal CST node of kind `ExperimentalPragma`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `ExperimentalPragma`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `ExperimentalPragma`.
     */
    cst: NonterminalNode);
    /**
     * Returns the child node that has the label `experimental_keyword`.
     */
    get experimentalKeyword(): TerminalNode;
    /**
     * Returns the child node that has the label `feature`.
     */
    get feature(): ExperimentalFeature;
}
/**
 * This node represents a `VersionPragma` nonterminal, with the following structure:
 *
 * ```ebnf
 * VersionPragma = (* solidity_keyword: *) SOLIDITY_KEYWORD
 *                 (* sets: *) VersionExpressionSets;
 * ```
 */
export declare class VersionPragma {
    /**
     * The underlying nonterminal CST node of kind `VersionPragma`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `VersionPragma`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `VersionPragma`.
     */
    cst: NonterminalNode);
    /**
     * Returns the child node that has the label `solidity_keyword`.
     */
    get solidityKeyword(): TerminalNode;
    /**
     * Returns the child node that has the label `sets`.
     */
    get sets(): VersionExpressionSets;
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
export declare class VersionRange {
    /**
     * The underlying nonterminal CST node of kind `VersionRange`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `VersionRange`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `VersionRange`.
     */
    cst: NonterminalNode);
    /**
     * Returns the child node that has the label `start`.
     */
    get start(): VersionLiteral;
    /**
     * Returns the child node that has the label `minus`.
     */
    get minus(): TerminalNode;
    /**
     * Returns the child node that has the label `end`.
     */
    get end(): VersionLiteral;
}
/**
 * This node represents a `VersionTerm` nonterminal, with the following structure:
 *
 * ```ebnf
 * VersionTerm = (* operator: *) VersionOperator?
 *               (* literal: *) VersionLiteral;
 * ```
 */
export declare class VersionTerm {
    /**
     * The underlying nonterminal CST node of kind `VersionTerm`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `VersionTerm`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `VersionTerm`.
     */
    cst: NonterminalNode);
    /**
     * Returns the child node that has the label `operator`.
     */
    get operator(): VersionOperator | undefined;
    /**
     * Returns the child node that has the label `literal`.
     */
    get literal(): VersionLiteral;
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
export declare class ImportDirective {
    /**
     * The underlying nonterminal CST node of kind `ImportDirective`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `ImportDirective`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `ImportDirective`.
     */
    cst: NonterminalNode);
    /**
     * Returns the child node that has the label `import_keyword`.
     */
    get importKeyword(): TerminalNode;
    /**
     * Returns the child node that has the label `clause`.
     */
    get clause(): ImportClause;
    /**
     * Returns the child node that has the label `semicolon`.
     */
    get semicolon(): TerminalNode;
}
/**
 * This node represents a `PathImport` nonterminal, with the following structure:
 *
 * ```ebnf
 * PathImport = (* path: *) StringLiteral
 *              (* alias: *) ImportAlias?;
 * ```
 */
export declare class PathImport {
    /**
     * The underlying nonterminal CST node of kind `PathImport`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `PathImport`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `PathImport`.
     */
    cst: NonterminalNode);
    /**
     * Returns the child node that has the label `path`.
     */
    get path(): StringLiteral;
    /**
     * Returns the child node that has the label `alias`.
     */
    get alias(): ImportAlias | undefined;
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
export declare class NamedImport {
    /**
     * The underlying nonterminal CST node of kind `NamedImport`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `NamedImport`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `NamedImport`.
     */
    cst: NonterminalNode);
    /**
     * Returns the child node that has the label `asterisk`.
     */
    get asterisk(): TerminalNode;
    /**
     * Returns the child node that has the label `alias`.
     */
    get alias(): ImportAlias;
    /**
     * Returns the child node that has the label `from_keyword`.
     */
    get fromKeyword(): TerminalNode;
    /**
     * Returns the child node that has the label `path`.
     */
    get path(): StringLiteral;
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
export declare class ImportDeconstruction {
    /**
     * The underlying nonterminal CST node of kind `ImportDeconstruction`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `ImportDeconstruction`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `ImportDeconstruction`.
     */
    cst: NonterminalNode);
    /**
     * Returns the child node that has the label `open_brace`.
     */
    get openBrace(): TerminalNode;
    /**
     * Returns the child node that has the label `symbols`.
     */
    get symbols(): ImportDeconstructionSymbols;
    /**
     * Returns the child node that has the label `close_brace`.
     */
    get closeBrace(): TerminalNode;
    /**
     * Returns the child node that has the label `from_keyword`.
     */
    get fromKeyword(): TerminalNode;
    /**
     * Returns the child node that has the label `path`.
     */
    get path(): StringLiteral;
}
/**
 * This node represents a `ImportDeconstructionSymbol` nonterminal, with the following structure:
 *
 * ```ebnf
 * ImportDeconstructionSymbol = (* name: *) IDENTIFIER
 *                              (* alias: *) ImportAlias?;
 * ```
 */
export declare class ImportDeconstructionSymbol {
    /**
     * The underlying nonterminal CST node of kind `ImportDeconstructionSymbol`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `ImportDeconstructionSymbol`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `ImportDeconstructionSymbol`.
     */
    cst: NonterminalNode);
    /**
     * Returns the child node that has the label `name`.
     */
    get name(): TerminalNode;
    /**
     * Returns the child node that has the label `alias`.
     */
    get alias(): ImportAlias | undefined;
}
/**
 * This node represents a `ImportAlias` nonterminal, with the following structure:
 *
 * ```ebnf
 * ImportAlias = (* as_keyword: *) AS_KEYWORD
 *               (* identifier: *) IDENTIFIER;
 * ```
 */
export declare class ImportAlias {
    /**
     * The underlying nonterminal CST node of kind `ImportAlias`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `ImportAlias`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `ImportAlias`.
     */
    cst: NonterminalNode);
    /**
     * Returns the child node that has the label `as_keyword`.
     */
    get asKeyword(): TerminalNode;
    /**
     * Returns the child node that has the label `identifier`.
     */
    get identifier(): TerminalNode;
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
export declare class UsingDirective {
    /**
     * The underlying nonterminal CST node of kind `UsingDirective`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `UsingDirective`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `UsingDirective`.
     */
    cst: NonterminalNode);
    /**
     * Returns the child node that has the label `using_keyword`.
     */
    get usingKeyword(): TerminalNode;
    /**
     * Returns the child node that has the label `clause`.
     */
    get clause(): UsingClause;
    /**
     * Returns the child node that has the label `for_keyword`.
     */
    get forKeyword(): TerminalNode;
    /**
     * Returns the child node that has the label `target`.
     */
    get target(): UsingTarget;
    /**
     * Returns the child node that has the label `global_keyword`.
     */
    get globalKeyword(): TerminalNode | undefined;
    /**
     * Returns the child node that has the label `semicolon`.
     */
    get semicolon(): TerminalNode;
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
export declare class UsingDeconstruction {
    /**
     * The underlying nonterminal CST node of kind `UsingDeconstruction`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `UsingDeconstruction`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `UsingDeconstruction`.
     */
    cst: NonterminalNode);
    /**
     * Returns the child node that has the label `open_brace`.
     */
    get openBrace(): TerminalNode;
    /**
     * Returns the child node that has the label `symbols`.
     */
    get symbols(): UsingDeconstructionSymbols;
    /**
     * Returns the child node that has the label `close_brace`.
     */
    get closeBrace(): TerminalNode;
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
export declare class UsingDeconstructionSymbol {
    /**
     * The underlying nonterminal CST node of kind `UsingDeconstructionSymbol`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `UsingDeconstructionSymbol`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `UsingDeconstructionSymbol`.
     */
    cst: NonterminalNode);
    /**
     * Returns the child node that has the label `name`.
     */
    get name(): IdentifierPath;
    /**
     * Returns the child node that has the label `alias`.
     */
    get alias(): UsingAlias | undefined;
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
export declare class UsingAlias {
    /**
     * The underlying nonterminal CST node of kind `UsingAlias`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `UsingAlias`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `UsingAlias`.
     */
    cst: NonterminalNode);
    /**
     * Returns the child node that has the label `as_keyword`.
     */
    get asKeyword(): TerminalNode;
    /**
     * Returns the child node that has the label `operator`.
     */
    get operator(): UsingOperator;
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
export declare class ContractDefinition {
    /**
     * The underlying nonterminal CST node of kind `ContractDefinition`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `ContractDefinition`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `ContractDefinition`.
     */
    cst: NonterminalNode);
    /**
     * Returns the child node that has the label `abstract_keyword`.
     */
    get abstractKeyword(): TerminalNode | undefined;
    /**
     * Returns the child node that has the label `contract_keyword`.
     */
    get contractKeyword(): TerminalNode;
    /**
     * Returns the child node that has the label `name`.
     */
    get name(): TerminalNode;
    /**
     * Returns the child node that has the label `specifiers`.
     */
    get specifiers(): ContractSpecifiers;
    /**
     * Returns the child node that has the label `open_brace`.
     */
    get openBrace(): TerminalNode;
    /**
     * Returns the child node that has the label `members`.
     */
    get members(): ContractMembers;
    /**
     * Returns the child node that has the label `close_brace`.
     */
    get closeBrace(): TerminalNode;
}
/**
 * This node represents a `InheritanceSpecifier` nonterminal, with the following structure:
 *
 * ```ebnf
 * InheritanceSpecifier = (* is_keyword: *) IS_KEYWORD
 *                        (* types: *) InheritanceTypes;
 * ```
 */
export declare class InheritanceSpecifier {
    /**
     * The underlying nonterminal CST node of kind `InheritanceSpecifier`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `InheritanceSpecifier`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `InheritanceSpecifier`.
     */
    cst: NonterminalNode);
    /**
     * Returns the child node that has the label `is_keyword`.
     */
    get isKeyword(): TerminalNode;
    /**
     * Returns the child node that has the label `types`.
     */
    get types(): InheritanceTypes;
}
/**
 * This node represents a `InheritanceType` nonterminal, with the following structure:
 *
 * ```ebnf
 * InheritanceType = (* type_name: *) IdentifierPath
 *                   (* arguments: *) ArgumentsDeclaration?;
 * ```
 */
export declare class InheritanceType {
    /**
     * The underlying nonterminal CST node of kind `InheritanceType`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `InheritanceType`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `InheritanceType`.
     */
    cst: NonterminalNode);
    /**
     * Returns the child node that has the label `type_name`.
     */
    get typeName(): IdentifierPath;
    /**
     * Returns the child node that has the label `arguments`.
     */
    get arguments(): ArgumentsDeclaration | undefined;
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
export declare class StorageLayoutSpecifier {
    /**
     * The underlying nonterminal CST node of kind `StorageLayoutSpecifier`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `StorageLayoutSpecifier`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `StorageLayoutSpecifier`.
     */
    cst: NonterminalNode);
    /**
     * Returns the child node that has the label `layout_keyword`.
     */
    get layoutKeyword(): TerminalNode;
    /**
     * Returns the child node that has the label `at_keyword`.
     */
    get atKeyword(): TerminalNode;
    /**
     * Returns the child node that has the label `expression`.
     */
    get expression(): Expression;
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
export declare class InterfaceDefinition {
    /**
     * The underlying nonterminal CST node of kind `InterfaceDefinition`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `InterfaceDefinition`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `InterfaceDefinition`.
     */
    cst: NonterminalNode);
    /**
     * Returns the child node that has the label `interface_keyword`.
     */
    get interfaceKeyword(): TerminalNode;
    /**
     * Returns the child node that has the label `name`.
     */
    get name(): TerminalNode;
    /**
     * Returns the child node that has the label `inheritance`.
     */
    get inheritance(): InheritanceSpecifier | undefined;
    /**
     * Returns the child node that has the label `open_brace`.
     */
    get openBrace(): TerminalNode;
    /**
     * Returns the child node that has the label `members`.
     */
    get members(): InterfaceMembers;
    /**
     * Returns the child node that has the label `close_brace`.
     */
    get closeBrace(): TerminalNode;
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
export declare class LibraryDefinition {
    /**
     * The underlying nonterminal CST node of kind `LibraryDefinition`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `LibraryDefinition`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `LibraryDefinition`.
     */
    cst: NonterminalNode);
    /**
     * Returns the child node that has the label `library_keyword`.
     */
    get libraryKeyword(): TerminalNode;
    /**
     * Returns the child node that has the label `name`.
     */
    get name(): TerminalNode;
    /**
     * Returns the child node that has the label `open_brace`.
     */
    get openBrace(): TerminalNode;
    /**
     * Returns the child node that has the label `members`.
     */
    get members(): LibraryMembers;
    /**
     * Returns the child node that has the label `close_brace`.
     */
    get closeBrace(): TerminalNode;
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
export declare class StructDefinition {
    /**
     * The underlying nonterminal CST node of kind `StructDefinition`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `StructDefinition`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `StructDefinition`.
     */
    cst: NonterminalNode);
    /**
     * Returns the child node that has the label `struct_keyword`.
     */
    get structKeyword(): TerminalNode;
    /**
     * Returns the child node that has the label `name`.
     */
    get name(): TerminalNode;
    /**
     * Returns the child node that has the label `open_brace`.
     */
    get openBrace(): TerminalNode;
    /**
     * Returns the child node that has the label `members`.
     */
    get members(): StructMembers;
    /**
     * Returns the child node that has the label `close_brace`.
     */
    get closeBrace(): TerminalNode;
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
export declare class StructMember {
    /**
     * The underlying nonterminal CST node of kind `StructMember`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `StructMember`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `StructMember`.
     */
    cst: NonterminalNode);
    /**
     * Returns the child node that has the label `type_name`.
     */
    get typeName(): TypeName;
    /**
     * Returns the child node that has the label `name`.
     */
    get name(): TerminalNode;
    /**
     * Returns the child node that has the label `semicolon`.
     */
    get semicolon(): TerminalNode;
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
export declare class EnumDefinition {
    /**
     * The underlying nonterminal CST node of kind `EnumDefinition`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `EnumDefinition`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `EnumDefinition`.
     */
    cst: NonterminalNode);
    /**
     * Returns the child node that has the label `enum_keyword`.
     */
    get enumKeyword(): TerminalNode;
    /**
     * Returns the child node that has the label `name`.
     */
    get name(): TerminalNode;
    /**
     * Returns the child node that has the label `open_brace`.
     */
    get openBrace(): TerminalNode;
    /**
     * Returns the child node that has the label `members`.
     */
    get members(): EnumMembers;
    /**
     * Returns the child node that has the label `close_brace`.
     */
    get closeBrace(): TerminalNode;
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
export declare class ConstantDefinition {
    /**
     * The underlying nonterminal CST node of kind `ConstantDefinition`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `ConstantDefinition`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `ConstantDefinition`.
     */
    cst: NonterminalNode);
    /**
     * Returns the child node that has the label `type_name`.
     */
    get typeName(): TypeName;
    /**
     * Returns the child node that has the label `constant_keyword`.
     */
    get constantKeyword(): TerminalNode;
    /**
     * Returns the child node that has the label `name`.
     */
    get name(): TerminalNode;
    /**
     * Returns the child node that has the label `equal`.
     */
    get equal(): TerminalNode;
    /**
     * Returns the child node that has the label `value`.
     */
    get value(): Expression;
    /**
     * Returns the child node that has the label `semicolon`.
     */
    get semicolon(): TerminalNode;
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
export declare class StateVariableDefinition {
    /**
     * The underlying nonterminal CST node of kind `StateVariableDefinition`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `StateVariableDefinition`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `StateVariableDefinition`.
     */
    cst: NonterminalNode);
    /**
     * Returns the child node that has the label `type_name`.
     */
    get typeName(): TypeName;
    /**
     * Returns the child node that has the label `attributes`.
     */
    get attributes(): StateVariableAttributes;
    /**
     * Returns the child node that has the label `name`.
     */
    get name(): TerminalNode;
    /**
     * Returns the child node that has the label `value`.
     */
    get value(): StateVariableDefinitionValue | undefined;
    /**
     * Returns the child node that has the label `semicolon`.
     */
    get semicolon(): TerminalNode;
}
/**
 * This node represents a `StateVariableDefinitionValue` nonterminal, with the following structure:
 *
 * ```ebnf
 * StateVariableDefinitionValue = (* equal: *) EQUAL
 *                                (* value: *) Expression;
 * ```
 */
export declare class StateVariableDefinitionValue {
    /**
     * The underlying nonterminal CST node of kind `StateVariableDefinitionValue`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `StateVariableDefinitionValue`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `StateVariableDefinitionValue`.
     */
    cst: NonterminalNode);
    /**
     * Returns the child node that has the label `equal`.
     */
    get equal(): TerminalNode;
    /**
     * Returns the child node that has the label `value`.
     */
    get value(): Expression;
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
export declare class FunctionDefinition {
    /**
     * The underlying nonterminal CST node of kind `FunctionDefinition`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `FunctionDefinition`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `FunctionDefinition`.
     */
    cst: NonterminalNode);
    /**
     * Returns the child node that has the label `function_keyword`.
     */
    get functionKeyword(): TerminalNode;
    /**
     * Returns the child node that has the label `name`.
     */
    get name(): FunctionName;
    /**
     * Returns the child node that has the label `parameters`.
     */
    get parameters(): ParametersDeclaration;
    /**
     * Returns the child node that has the label `attributes`.
     */
    get attributes(): FunctionAttributes;
    /**
     * Returns the child node that has the label `returns`.
     */
    get returns(): ReturnsDeclaration | undefined;
    /**
     * Returns the child node that has the label `body`.
     */
    get body(): FunctionBody;
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
export declare class ParametersDeclaration {
    /**
     * The underlying nonterminal CST node of kind `ParametersDeclaration`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `ParametersDeclaration`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `ParametersDeclaration`.
     */
    cst: NonterminalNode);
    /**
     * Returns the child node that has the label `open_paren`.
     */
    get openParen(): TerminalNode;
    /**
     * Returns the child node that has the label `parameters`.
     */
    get parameters(): Parameters;
    /**
     * Returns the child node that has the label `close_paren`.
     */
    get closeParen(): TerminalNode;
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
export declare class Parameter {
    /**
     * The underlying nonterminal CST node of kind `Parameter`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `Parameter`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `Parameter`.
     */
    cst: NonterminalNode);
    /**
     * Returns the child node that has the label `type_name`.
     */
    get typeName(): TypeName;
    /**
     * Returns the child node that has the label `storage_location`.
     */
    get storageLocation(): StorageLocation | undefined;
    /**
     * Returns the child node that has the label `name`.
     */
    get name(): TerminalNode | undefined;
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
export declare class OverrideSpecifier {
    /**
     * The underlying nonterminal CST node of kind `OverrideSpecifier`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `OverrideSpecifier`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `OverrideSpecifier`.
     */
    cst: NonterminalNode);
    /**
     * Returns the child node that has the label `override_keyword`.
     */
    get overrideKeyword(): TerminalNode;
    /**
     * Returns the child node that has the label `overridden`.
     */
    get overridden(): OverridePathsDeclaration | undefined;
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
export declare class OverridePathsDeclaration {
    /**
     * The underlying nonterminal CST node of kind `OverridePathsDeclaration`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `OverridePathsDeclaration`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `OverridePathsDeclaration`.
     */
    cst: NonterminalNode);
    /**
     * Returns the child node that has the label `open_paren`.
     */
    get openParen(): TerminalNode;
    /**
     * Returns the child node that has the label `paths`.
     */
    get paths(): OverridePaths;
    /**
     * Returns the child node that has the label `close_paren`.
     */
    get closeParen(): TerminalNode;
}
/**
 * This node represents a `ReturnsDeclaration` nonterminal, with the following structure:
 *
 * ```ebnf
 * ReturnsDeclaration = (* returns_keyword: *) RETURNS_KEYWORD
 *                      (* variables: *) ParametersDeclaration;
 * ```
 */
export declare class ReturnsDeclaration {
    /**
     * The underlying nonterminal CST node of kind `ReturnsDeclaration`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `ReturnsDeclaration`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `ReturnsDeclaration`.
     */
    cst: NonterminalNode);
    /**
     * Returns the child node that has the label `returns_keyword`.
     */
    get returnsKeyword(): TerminalNode;
    /**
     * Returns the child node that has the label `variables`.
     */
    get variables(): ParametersDeclaration;
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
export declare class ConstructorDefinition {
    /**
     * The underlying nonterminal CST node of kind `ConstructorDefinition`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `ConstructorDefinition`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `ConstructorDefinition`.
     */
    cst: NonterminalNode);
    /**
     * Returns the child node that has the label `constructor_keyword`.
     */
    get constructorKeyword(): TerminalNode;
    /**
     * Returns the child node that has the label `parameters`.
     */
    get parameters(): ParametersDeclaration;
    /**
     * Returns the child node that has the label `attributes`.
     */
    get attributes(): ConstructorAttributes;
    /**
     * Returns the child node that has the label `body`.
     */
    get body(): Block;
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
export declare class UnnamedFunctionDefinition {
    /**
     * The underlying nonterminal CST node of kind `UnnamedFunctionDefinition`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `UnnamedFunctionDefinition`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `UnnamedFunctionDefinition`.
     */
    cst: NonterminalNode);
    /**
     * Returns the child node that has the label `function_keyword`.
     */
    get functionKeyword(): TerminalNode;
    /**
     * Returns the child node that has the label `parameters`.
     */
    get parameters(): ParametersDeclaration;
    /**
     * Returns the child node that has the label `attributes`.
     */
    get attributes(): UnnamedFunctionAttributes;
    /**
     * Returns the child node that has the label `body`.
     */
    get body(): FunctionBody;
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
export declare class FallbackFunctionDefinition {
    /**
     * The underlying nonterminal CST node of kind `FallbackFunctionDefinition`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `FallbackFunctionDefinition`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `FallbackFunctionDefinition`.
     */
    cst: NonterminalNode);
    /**
     * Returns the child node that has the label `fallback_keyword`.
     */
    get fallbackKeyword(): TerminalNode;
    /**
     * Returns the child node that has the label `parameters`.
     */
    get parameters(): ParametersDeclaration;
    /**
     * Returns the child node that has the label `attributes`.
     */
    get attributes(): FallbackFunctionAttributes;
    /**
     * Returns the child node that has the label `returns`.
     */
    get returns(): ReturnsDeclaration | undefined;
    /**
     * Returns the child node that has the label `body`.
     */
    get body(): FunctionBody;
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
export declare class ReceiveFunctionDefinition {
    /**
     * The underlying nonterminal CST node of kind `ReceiveFunctionDefinition`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `ReceiveFunctionDefinition`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `ReceiveFunctionDefinition`.
     */
    cst: NonterminalNode);
    /**
     * Returns the child node that has the label `receive_keyword`.
     */
    get receiveKeyword(): TerminalNode;
    /**
     * Returns the child node that has the label `parameters`.
     */
    get parameters(): ParametersDeclaration;
    /**
     * Returns the child node that has the label `attributes`.
     */
    get attributes(): ReceiveFunctionAttributes;
    /**
     * Returns the child node that has the label `body`.
     */
    get body(): FunctionBody;
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
export declare class ModifierDefinition {
    /**
     * The underlying nonterminal CST node of kind `ModifierDefinition`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `ModifierDefinition`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `ModifierDefinition`.
     */
    cst: NonterminalNode);
    /**
     * Returns the child node that has the label `modifier_keyword`.
     */
    get modifierKeyword(): TerminalNode;
    /**
     * Returns the child node that has the label `name`.
     */
    get name(): TerminalNode;
    /**
     * Returns the child node that has the label `parameters`.
     */
    get parameters(): ParametersDeclaration | undefined;
    /**
     * Returns the child node that has the label `attributes`.
     */
    get attributes(): ModifierAttributes;
    /**
     * Returns the child node that has the label `body`.
     */
    get body(): FunctionBody;
}
/**
 * This node represents a `ModifierInvocation` nonterminal, with the following structure:
 *
 * ```ebnf
 * ModifierInvocation = (* name: *) IdentifierPath
 *                      (* arguments: *) ArgumentsDeclaration?;
 * ```
 */
export declare class ModifierInvocation {
    /**
     * The underlying nonterminal CST node of kind `ModifierInvocation`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `ModifierInvocation`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `ModifierInvocation`.
     */
    cst: NonterminalNode);
    /**
     * Returns the child node that has the label `name`.
     */
    get name(): IdentifierPath;
    /**
     * Returns the child node that has the label `arguments`.
     */
    get arguments(): ArgumentsDeclaration | undefined;
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
export declare class EventDefinition {
    /**
     * The underlying nonterminal CST node of kind `EventDefinition`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `EventDefinition`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `EventDefinition`.
     */
    cst: NonterminalNode);
    /**
     * Returns the child node that has the label `event_keyword`.
     */
    get eventKeyword(): TerminalNode;
    /**
     * Returns the child node that has the label `name`.
     */
    get name(): TerminalNode;
    /**
     * Returns the child node that has the label `parameters`.
     */
    get parameters(): EventParametersDeclaration;
    /**
     * Returns the child node that has the label `anonymous_keyword`.
     */
    get anonymousKeyword(): TerminalNode | undefined;
    /**
     * Returns the child node that has the label `semicolon`.
     */
    get semicolon(): TerminalNode;
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
export declare class EventParametersDeclaration {
    /**
     * The underlying nonterminal CST node of kind `EventParametersDeclaration`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `EventParametersDeclaration`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `EventParametersDeclaration`.
     */
    cst: NonterminalNode);
    /**
     * Returns the child node that has the label `open_paren`.
     */
    get openParen(): TerminalNode;
    /**
     * Returns the child node that has the label `parameters`.
     */
    get parameters(): EventParameters;
    /**
     * Returns the child node that has the label `close_paren`.
     */
    get closeParen(): TerminalNode;
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
export declare class EventParameter {
    /**
     * The underlying nonterminal CST node of kind `EventParameter`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `EventParameter`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `EventParameter`.
     */
    cst: NonterminalNode);
    /**
     * Returns the child node that has the label `type_name`.
     */
    get typeName(): TypeName;
    /**
     * Returns the child node that has the label `indexed_keyword`.
     */
    get indexedKeyword(): TerminalNode | undefined;
    /**
     * Returns the child node that has the label `name`.
     */
    get name(): TerminalNode | undefined;
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
export declare class UserDefinedValueTypeDefinition {
    /**
     * The underlying nonterminal CST node of kind `UserDefinedValueTypeDefinition`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `UserDefinedValueTypeDefinition`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `UserDefinedValueTypeDefinition`.
     */
    cst: NonterminalNode);
    /**
     * Returns the child node that has the label `type_keyword`.
     */
    get typeKeyword(): TerminalNode;
    /**
     * Returns the child node that has the label `name`.
     */
    get name(): TerminalNode;
    /**
     * Returns the child node that has the label `is_keyword`.
     */
    get isKeyword(): TerminalNode;
    /**
     * Returns the child node that has the label `value_type`.
     */
    get valueType(): ElementaryType;
    /**
     * Returns the child node that has the label `semicolon`.
     */
    get semicolon(): TerminalNode;
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
export declare class ErrorDefinition {
    /**
     * The underlying nonterminal CST node of kind `ErrorDefinition`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `ErrorDefinition`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `ErrorDefinition`.
     */
    cst: NonterminalNode);
    /**
     * Returns the child node that has the label `error_keyword`.
     */
    get errorKeyword(): TerminalNode;
    /**
     * Returns the child node that has the label `name`.
     */
    get name(): TerminalNode;
    /**
     * Returns the child node that has the label `members`.
     */
    get members(): ErrorParametersDeclaration;
    /**
     * Returns the child node that has the label `semicolon`.
     */
    get semicolon(): TerminalNode;
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
export declare class ErrorParametersDeclaration {
    /**
     * The underlying nonterminal CST node of kind `ErrorParametersDeclaration`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `ErrorParametersDeclaration`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `ErrorParametersDeclaration`.
     */
    cst: NonterminalNode);
    /**
     * Returns the child node that has the label `open_paren`.
     */
    get openParen(): TerminalNode;
    /**
     * Returns the child node that has the label `parameters`.
     */
    get parameters(): ErrorParameters;
    /**
     * Returns the child node that has the label `close_paren`.
     */
    get closeParen(): TerminalNode;
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
export declare class ErrorParameter {
    /**
     * The underlying nonterminal CST node of kind `ErrorParameter`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `ErrorParameter`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `ErrorParameter`.
     */
    cst: NonterminalNode);
    /**
     * Returns the child node that has the label `type_name`.
     */
    get typeName(): TypeName;
    /**
     * Returns the child node that has the label `name`.
     */
    get name(): TerminalNode | undefined;
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
export declare class ArrayTypeName {
    /**
     * The underlying nonterminal CST node of kind `ArrayTypeName`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `ArrayTypeName`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `ArrayTypeName`.
     */
    cst: NonterminalNode);
    /**
     * Returns the child node that has the label `operand`.
     */
    get operand(): TypeName;
    /**
     * Returns the child node that has the label `open_bracket`.
     */
    get openBracket(): TerminalNode;
    /**
     * Returns the child node that has the label `index`.
     */
    get index(): Expression | undefined;
    /**
     * Returns the child node that has the label `close_bracket`.
     */
    get closeBracket(): TerminalNode;
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
export declare class FunctionType {
    /**
     * The underlying nonterminal CST node of kind `FunctionType`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `FunctionType`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `FunctionType`.
     */
    cst: NonterminalNode);
    /**
     * Returns the child node that has the label `function_keyword`.
     */
    get functionKeyword(): TerminalNode;
    /**
     * Returns the child node that has the label `parameters`.
     */
    get parameters(): ParametersDeclaration;
    /**
     * Returns the child node that has the label `attributes`.
     */
    get attributes(): FunctionTypeAttributes;
    /**
     * Returns the child node that has the label `returns`.
     */
    get returns(): ReturnsDeclaration | undefined;
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
export declare class MappingType {
    /**
     * The underlying nonterminal CST node of kind `MappingType`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `MappingType`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `MappingType`.
     */
    cst: NonterminalNode);
    /**
     * Returns the child node that has the label `mapping_keyword`.
     */
    get mappingKeyword(): TerminalNode;
    /**
     * Returns the child node that has the label `open_paren`.
     */
    get openParen(): TerminalNode;
    /**
     * Returns the child node that has the label `key_type`.
     */
    get keyType(): MappingKey;
    /**
     * Returns the child node that has the label `equal_greater_than`.
     */
    get equalGreaterThan(): TerminalNode;
    /**
     * Returns the child node that has the label `value_type`.
     */
    get valueType(): MappingValue;
    /**
     * Returns the child node that has the label `close_paren`.
     */
    get closeParen(): TerminalNode;
}
/**
 * This node represents a `MappingKey` nonterminal, with the following structure:
 *
 * ```ebnf
 * MappingKey = (* key_type: *) MappingKeyType
 *              (* name: *) IDENTIFIER?; (* Introduced in 0.8.18 *)
 * ```
 */
export declare class MappingKey {
    /**
     * The underlying nonterminal CST node of kind `MappingKey`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `MappingKey`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `MappingKey`.
     */
    cst: NonterminalNode);
    /**
     * Returns the child node that has the label `key_type`.
     */
    get keyType(): MappingKeyType;
    /**
     * Returns the child node that has the label `name`.
     */
    get name(): TerminalNode | undefined;
}
/**
 * This node represents a `MappingValue` nonterminal, with the following structure:
 *
 * ```ebnf
 * MappingValue = (* type_name: *) TypeName
 *                (* name: *) IDENTIFIER?; (* Introduced in 0.8.18 *)
 * ```
 */
export declare class MappingValue {
    /**
     * The underlying nonterminal CST node of kind `MappingValue`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `MappingValue`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `MappingValue`.
     */
    cst: NonterminalNode);
    /**
     * Returns the child node that has the label `type_name`.
     */
    get typeName(): TypeName;
    /**
     * Returns the child node that has the label `name`.
     */
    get name(): TerminalNode | undefined;
}
/**
 * This node represents a `AddressType` nonterminal, with the following structure:
 *
 * ```ebnf
 * AddressType = (* address_keyword: *) ADDRESS_KEYWORD
 *               (* payable_keyword: *) PAYABLE_KEYWORD?; (* Introduced in 0.5.0 *)
 * ```
 */
export declare class AddressType {
    /**
     * The underlying nonterminal CST node of kind `AddressType`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `AddressType`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `AddressType`.
     */
    cst: NonterminalNode);
    /**
     * Returns the child node that has the label `address_keyword`.
     */
    get addressKeyword(): TerminalNode;
    /**
     * Returns the child node that has the label `payable_keyword`.
     */
    get payableKeyword(): TerminalNode | undefined;
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
export declare class Block {
    /**
     * The underlying nonterminal CST node of kind `Block`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `Block`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `Block`.
     */
    cst: NonterminalNode);
    /**
     * Returns the child node that has the label `open_brace`.
     */
    get openBrace(): TerminalNode;
    /**
     * Returns the child node that has the label `statements`.
     */
    get statements(): Statements;
    /**
     * Returns the child node that has the label `close_brace`.
     */
    get closeBrace(): TerminalNode;
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
export declare class UncheckedBlock {
    /**
     * The underlying nonterminal CST node of kind `UncheckedBlock`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `UncheckedBlock`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `UncheckedBlock`.
     */
    cst: NonterminalNode);
    /**
     * Returns the child node that has the label `unchecked_keyword`.
     */
    get uncheckedKeyword(): TerminalNode;
    /**
     * Returns the child node that has the label `block`.
     */
    get block(): Block;
}
/**
 * This node represents a `ExpressionStatement` nonterminal, with the following structure:
 *
 * ```ebnf
 * ExpressionStatement = (* expression: *) Expression
 *                       (* semicolon: *) SEMICOLON;
 * ```
 */
export declare class ExpressionStatement {
    /**
     * The underlying nonterminal CST node of kind `ExpressionStatement`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `ExpressionStatement`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `ExpressionStatement`.
     */
    cst: NonterminalNode);
    /**
     * Returns the child node that has the label `expression`.
     */
    get expression(): Expression;
    /**
     * Returns the child node that has the label `semicolon`.
     */
    get semicolon(): TerminalNode;
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
export declare class AssemblyStatement {
    /**
     * The underlying nonterminal CST node of kind `AssemblyStatement`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `AssemblyStatement`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `AssemblyStatement`.
     */
    cst: NonterminalNode);
    /**
     * Returns the child node that has the label `assembly_keyword`.
     */
    get assemblyKeyword(): TerminalNode;
    /**
     * Returns the child node that has the label `label`.
     */
    get label(): StringLiteral | undefined;
    /**
     * Returns the child node that has the label `flags`.
     */
    get flags(): AssemblyFlagsDeclaration | undefined;
    /**
     * Returns the child node that has the label `body`.
     */
    get body(): YulBlock;
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
export declare class AssemblyFlagsDeclaration {
    /**
     * The underlying nonterminal CST node of kind `AssemblyFlagsDeclaration`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `AssemblyFlagsDeclaration`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `AssemblyFlagsDeclaration`.
     */
    cst: NonterminalNode);
    /**
     * Returns the child node that has the label `open_paren`.
     */
    get openParen(): TerminalNode;
    /**
     * Returns the child node that has the label `flags`.
     */
    get flags(): AssemblyFlags;
    /**
     * Returns the child node that has the label `close_paren`.
     */
    get closeParen(): TerminalNode;
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
export declare class TupleDeconstructionStatement {
    /**
     * The underlying nonterminal CST node of kind `TupleDeconstructionStatement`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `TupleDeconstructionStatement`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `TupleDeconstructionStatement`.
     */
    cst: NonterminalNode);
    /**
     * Returns the child node that has the label `var_keyword`.
     */
    get varKeyword(): TerminalNode | undefined;
    /**
     * Returns the child node that has the label `open_paren`.
     */
    get openParen(): TerminalNode;
    /**
     * Returns the child node that has the label `elements`.
     */
    get elements(): TupleDeconstructionElements;
    /**
     * Returns the child node that has the label `close_paren`.
     */
    get closeParen(): TerminalNode;
    /**
     * Returns the child node that has the label `equal`.
     */
    get equal(): TerminalNode;
    /**
     * Returns the child node that has the label `expression`.
     */
    get expression(): Expression;
    /**
     * Returns the child node that has the label `semicolon`.
     */
    get semicolon(): TerminalNode;
}
/**
 * This node represents a `TupleDeconstructionElement` nonterminal, with the following structure:
 *
 * ```ebnf
 * TupleDeconstructionElement = (* member: *) TupleMember?;
 * ```
 */
export declare class TupleDeconstructionElement {
    /**
     * The underlying nonterminal CST node of kind `TupleDeconstructionElement`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `TupleDeconstructionElement`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `TupleDeconstructionElement`.
     */
    cst: NonterminalNode);
    /**
     * Returns the child node that has the label `member`.
     */
    get member(): TupleMember | undefined;
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
export declare class TypedTupleMember {
    /**
     * The underlying nonterminal CST node of kind `TypedTupleMember`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `TypedTupleMember`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `TypedTupleMember`.
     */
    cst: NonterminalNode);
    /**
     * Returns the child node that has the label `type_name`.
     */
    get typeName(): TypeName;
    /**
     * Returns the child node that has the label `storage_location`.
     */
    get storageLocation(): StorageLocation | undefined;
    /**
     * Returns the child node that has the label `name`.
     */
    get name(): TerminalNode;
}
/**
 * This node represents a `UntypedTupleMember` nonterminal, with the following structure:
 *
 * ```ebnf
 * UntypedTupleMember = (* storage_location: *) StorageLocation?
 *                      (* name: *) IDENTIFIER;
 * ```
 */
export declare class UntypedTupleMember {
    /**
     * The underlying nonterminal CST node of kind `UntypedTupleMember`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `UntypedTupleMember`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `UntypedTupleMember`.
     */
    cst: NonterminalNode);
    /**
     * Returns the child node that has the label `storage_location`.
     */
    get storageLocation(): StorageLocation | undefined;
    /**
     * Returns the child node that has the label `name`.
     */
    get name(): TerminalNode;
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
export declare class VariableDeclarationStatement {
    /**
     * The underlying nonterminal CST node of kind `VariableDeclarationStatement`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `VariableDeclarationStatement`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `VariableDeclarationStatement`.
     */
    cst: NonterminalNode);
    /**
     * Returns the child node that has the label `variable_type`.
     */
    get variableType(): VariableDeclarationType;
    /**
     * Returns the child node that has the label `storage_location`.
     */
    get storageLocation(): StorageLocation | undefined;
    /**
     * Returns the child node that has the label `name`.
     */
    get name(): TerminalNode;
    /**
     * Returns the child node that has the label `value`.
     */
    get value(): VariableDeclarationValue | undefined;
    /**
     * Returns the child node that has the label `semicolon`.
     */
    get semicolon(): TerminalNode;
}
/**
 * This node represents a `VariableDeclarationValue` nonterminal, with the following structure:
 *
 * ```ebnf
 * VariableDeclarationValue = (* equal: *) EQUAL
 *                            (* expression: *) Expression;
 * ```
 */
export declare class VariableDeclarationValue {
    /**
     * The underlying nonterminal CST node of kind `VariableDeclarationValue`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `VariableDeclarationValue`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `VariableDeclarationValue`.
     */
    cst: NonterminalNode);
    /**
     * Returns the child node that has the label `equal`.
     */
    get equal(): TerminalNode;
    /**
     * Returns the child node that has the label `expression`.
     */
    get expression(): Expression;
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
export declare class IfStatement {
    /**
     * The underlying nonterminal CST node of kind `IfStatement`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `IfStatement`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `IfStatement`.
     */
    cst: NonterminalNode);
    /**
     * Returns the child node that has the label `if_keyword`.
     */
    get ifKeyword(): TerminalNode;
    /**
     * Returns the child node that has the label `open_paren`.
     */
    get openParen(): TerminalNode;
    /**
     * Returns the child node that has the label `condition`.
     */
    get condition(): Expression;
    /**
     * Returns the child node that has the label `close_paren`.
     */
    get closeParen(): TerminalNode;
    /**
     * Returns the child node that has the label `body`.
     */
    get body(): Statement;
    /**
     * Returns the child node that has the label `else_branch`.
     */
    get elseBranch(): ElseBranch | undefined;
}
/**
 * This node represents a `ElseBranch` nonterminal, with the following structure:
 *
 * ```ebnf
 * ElseBranch = (* else_keyword: *) ELSE_KEYWORD
 *              (* body: *) Statement;
 * ```
 */
export declare class ElseBranch {
    /**
     * The underlying nonterminal CST node of kind `ElseBranch`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `ElseBranch`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `ElseBranch`.
     */
    cst: NonterminalNode);
    /**
     * Returns the child node that has the label `else_keyword`.
     */
    get elseKeyword(): TerminalNode;
    /**
     * Returns the child node that has the label `body`.
     */
    get body(): Statement;
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
export declare class ForStatement {
    /**
     * The underlying nonterminal CST node of kind `ForStatement`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `ForStatement`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `ForStatement`.
     */
    cst: NonterminalNode);
    /**
     * Returns the child node that has the label `for_keyword`.
     */
    get forKeyword(): TerminalNode;
    /**
     * Returns the child node that has the label `open_paren`.
     */
    get openParen(): TerminalNode;
    /**
     * Returns the child node that has the label `initialization`.
     */
    get initialization(): ForStatementInitialization;
    /**
     * Returns the child node that has the label `condition`.
     */
    get condition(): ForStatementCondition;
    /**
     * Returns the child node that has the label `iterator`.
     */
    get iterator(): Expression | undefined;
    /**
     * Returns the child node that has the label `close_paren`.
     */
    get closeParen(): TerminalNode;
    /**
     * Returns the child node that has the label `body`.
     */
    get body(): Statement;
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
export declare class WhileStatement {
    /**
     * The underlying nonterminal CST node of kind `WhileStatement`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `WhileStatement`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `WhileStatement`.
     */
    cst: NonterminalNode);
    /**
     * Returns the child node that has the label `while_keyword`.
     */
    get whileKeyword(): TerminalNode;
    /**
     * Returns the child node that has the label `open_paren`.
     */
    get openParen(): TerminalNode;
    /**
     * Returns the child node that has the label `condition`.
     */
    get condition(): Expression;
    /**
     * Returns the child node that has the label `close_paren`.
     */
    get closeParen(): TerminalNode;
    /**
     * Returns the child node that has the label `body`.
     */
    get body(): Statement;
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
export declare class DoWhileStatement {
    /**
     * The underlying nonterminal CST node of kind `DoWhileStatement`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `DoWhileStatement`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `DoWhileStatement`.
     */
    cst: NonterminalNode);
    /**
     * Returns the child node that has the label `do_keyword`.
     */
    get doKeyword(): TerminalNode;
    /**
     * Returns the child node that has the label `body`.
     */
    get body(): Statement;
    /**
     * Returns the child node that has the label `while_keyword`.
     */
    get whileKeyword(): TerminalNode;
    /**
     * Returns the child node that has the label `open_paren`.
     */
    get openParen(): TerminalNode;
    /**
     * Returns the child node that has the label `condition`.
     */
    get condition(): Expression;
    /**
     * Returns the child node that has the label `close_paren`.
     */
    get closeParen(): TerminalNode;
    /**
     * Returns the child node that has the label `semicolon`.
     */
    get semicolon(): TerminalNode;
}
/**
 * This node represents a `ContinueStatement` nonterminal, with the following structure:
 *
 * ```ebnf
 * ContinueStatement = (* continue_keyword: *) CONTINUE_KEYWORD
 *                     (* semicolon: *) SEMICOLON;
 * ```
 */
export declare class ContinueStatement {
    /**
     * The underlying nonterminal CST node of kind `ContinueStatement`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `ContinueStatement`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `ContinueStatement`.
     */
    cst: NonterminalNode);
    /**
     * Returns the child node that has the label `continue_keyword`.
     */
    get continueKeyword(): TerminalNode;
    /**
     * Returns the child node that has the label `semicolon`.
     */
    get semicolon(): TerminalNode;
}
/**
 * This node represents a `BreakStatement` nonterminal, with the following structure:
 *
 * ```ebnf
 * BreakStatement = (* break_keyword: *) BREAK_KEYWORD
 *                  (* semicolon: *) SEMICOLON;
 * ```
 */
export declare class BreakStatement {
    /**
     * The underlying nonterminal CST node of kind `BreakStatement`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `BreakStatement`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `BreakStatement`.
     */
    cst: NonterminalNode);
    /**
     * Returns the child node that has the label `break_keyword`.
     */
    get breakKeyword(): TerminalNode;
    /**
     * Returns the child node that has the label `semicolon`.
     */
    get semicolon(): TerminalNode;
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
export declare class ReturnStatement {
    /**
     * The underlying nonterminal CST node of kind `ReturnStatement`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `ReturnStatement`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `ReturnStatement`.
     */
    cst: NonterminalNode);
    /**
     * Returns the child node that has the label `return_keyword`.
     */
    get returnKeyword(): TerminalNode;
    /**
     * Returns the child node that has the label `expression`.
     */
    get expression(): Expression | undefined;
    /**
     * Returns the child node that has the label `semicolon`.
     */
    get semicolon(): TerminalNode;
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
export declare class EmitStatement {
    /**
     * The underlying nonterminal CST node of kind `EmitStatement`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `EmitStatement`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `EmitStatement`.
     */
    cst: NonterminalNode);
    /**
     * Returns the child node that has the label `emit_keyword`.
     */
    get emitKeyword(): TerminalNode;
    /**
     * Returns the child node that has the label `event`.
     */
    get event(): IdentifierPath;
    /**
     * Returns the child node that has the label `arguments`.
     */
    get arguments(): ArgumentsDeclaration;
    /**
     * Returns the child node that has the label `semicolon`.
     */
    get semicolon(): TerminalNode;
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
export declare class TryStatement {
    /**
     * The underlying nonterminal CST node of kind `TryStatement`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `TryStatement`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `TryStatement`.
     */
    cst: NonterminalNode);
    /**
     * Returns the child node that has the label `try_keyword`.
     */
    get tryKeyword(): TerminalNode;
    /**
     * Returns the child node that has the label `expression`.
     */
    get expression(): Expression;
    /**
     * Returns the child node that has the label `returns`.
     */
    get returns(): ReturnsDeclaration | undefined;
    /**
     * Returns the child node that has the label `body`.
     */
    get body(): Block;
    /**
     * Returns the child node that has the label `catch_clauses`.
     */
    get catchClauses(): CatchClauses;
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
export declare class CatchClause {
    /**
     * The underlying nonterminal CST node of kind `CatchClause`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `CatchClause`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `CatchClause`.
     */
    cst: NonterminalNode);
    /**
     * Returns the child node that has the label `catch_keyword`.
     */
    get catchKeyword(): TerminalNode;
    /**
     * Returns the child node that has the label `error`.
     */
    get error(): CatchClauseError | undefined;
    /**
     * Returns the child node that has the label `body`.
     */
    get body(): Block;
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
export declare class CatchClauseError {
    /**
     * The underlying nonterminal CST node of kind `CatchClauseError`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `CatchClauseError`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `CatchClauseError`.
     */
    cst: NonterminalNode);
    /**
     * Returns the child node that has the label `name`.
     */
    get name(): TerminalNode | undefined;
    /**
     * Returns the child node that has the label `parameters`.
     */
    get parameters(): ParametersDeclaration;
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
export declare class RevertStatement {
    /**
     * The underlying nonterminal CST node of kind `RevertStatement`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `RevertStatement`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `RevertStatement`.
     */
    cst: NonterminalNode);
    /**
     * Returns the child node that has the label `revert_keyword`.
     */
    get revertKeyword(): TerminalNode;
    /**
     * Returns the child node that has the label `error`.
     */
    get error(): IdentifierPath;
    /**
     * Returns the child node that has the label `arguments`.
     */
    get arguments(): ArgumentsDeclaration;
    /**
     * Returns the child node that has the label `semicolon`.
     */
    get semicolon(): TerminalNode;
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
export declare class ThrowStatement {
    /**
     * The underlying nonterminal CST node of kind `ThrowStatement`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `ThrowStatement`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `ThrowStatement`.
     */
    cst: NonterminalNode);
    /**
     * Returns the child node that has the label `throw_keyword`.
     */
    get throwKeyword(): TerminalNode;
    /**
     * Returns the child node that has the label `semicolon`.
     */
    get semicolon(): TerminalNode;
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
export declare class AssignmentExpression {
    /**
     * The underlying nonterminal CST node of kind `AssignmentExpression`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `AssignmentExpression`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `AssignmentExpression`.
     */
    cst: NonterminalNode);
    /**
     * Returns the child node that has the label `left_operand`.
     */
    get leftOperand(): Expression;
    /**
     * Returns the child node that has the label `operator`.
     */
    get operator(): TerminalNode;
    /**
     * Returns the child node that has the label `right_operand`.
     */
    get rightOperand(): Expression;
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
export declare class ConditionalExpression {
    /**
     * The underlying nonterminal CST node of kind `ConditionalExpression`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `ConditionalExpression`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `ConditionalExpression`.
     */
    cst: NonterminalNode);
    /**
     * Returns the child node that has the label `operand`.
     */
    get operand(): Expression;
    /**
     * Returns the child node that has the label `question_mark`.
     */
    get questionMark(): TerminalNode;
    /**
     * Returns the child node that has the label `true_expression`.
     */
    get trueExpression(): Expression;
    /**
     * Returns the child node that has the label `colon`.
     */
    get colon(): TerminalNode;
    /**
     * Returns the child node that has the label `false_expression`.
     */
    get falseExpression(): Expression;
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
export declare class OrExpression {
    /**
     * The underlying nonterminal CST node of kind `OrExpression`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `OrExpression`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `OrExpression`.
     */
    cst: NonterminalNode);
    /**
     * Returns the child node that has the label `left_operand`.
     */
    get leftOperand(): Expression;
    /**
     * Returns the child node that has the label `operator`.
     */
    get operator(): TerminalNode;
    /**
     * Returns the child node that has the label `right_operand`.
     */
    get rightOperand(): Expression;
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
export declare class AndExpression {
    /**
     * The underlying nonterminal CST node of kind `AndExpression`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `AndExpression`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `AndExpression`.
     */
    cst: NonterminalNode);
    /**
     * Returns the child node that has the label `left_operand`.
     */
    get leftOperand(): Expression;
    /**
     * Returns the child node that has the label `operator`.
     */
    get operator(): TerminalNode;
    /**
     * Returns the child node that has the label `right_operand`.
     */
    get rightOperand(): Expression;
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
export declare class EqualityExpression {
    /**
     * The underlying nonterminal CST node of kind `EqualityExpression`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `EqualityExpression`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `EqualityExpression`.
     */
    cst: NonterminalNode);
    /**
     * Returns the child node that has the label `left_operand`.
     */
    get leftOperand(): Expression;
    /**
     * Returns the child node that has the label `operator`.
     */
    get operator(): TerminalNode;
    /**
     * Returns the child node that has the label `right_operand`.
     */
    get rightOperand(): Expression;
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
export declare class InequalityExpression {
    /**
     * The underlying nonterminal CST node of kind `InequalityExpression`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `InequalityExpression`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `InequalityExpression`.
     */
    cst: NonterminalNode);
    /**
     * Returns the child node that has the label `left_operand`.
     */
    get leftOperand(): Expression;
    /**
     * Returns the child node that has the label `operator`.
     */
    get operator(): TerminalNode;
    /**
     * Returns the child node that has the label `right_operand`.
     */
    get rightOperand(): Expression;
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
export declare class BitwiseOrExpression {
    /**
     * The underlying nonterminal CST node of kind `BitwiseOrExpression`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `BitwiseOrExpression`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `BitwiseOrExpression`.
     */
    cst: NonterminalNode);
    /**
     * Returns the child node that has the label `left_operand`.
     */
    get leftOperand(): Expression;
    /**
     * Returns the child node that has the label `operator`.
     */
    get operator(): TerminalNode;
    /**
     * Returns the child node that has the label `right_operand`.
     */
    get rightOperand(): Expression;
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
export declare class BitwiseXorExpression {
    /**
     * The underlying nonterminal CST node of kind `BitwiseXorExpression`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `BitwiseXorExpression`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `BitwiseXorExpression`.
     */
    cst: NonterminalNode);
    /**
     * Returns the child node that has the label `left_operand`.
     */
    get leftOperand(): Expression;
    /**
     * Returns the child node that has the label `operator`.
     */
    get operator(): TerminalNode;
    /**
     * Returns the child node that has the label `right_operand`.
     */
    get rightOperand(): Expression;
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
export declare class BitwiseAndExpression {
    /**
     * The underlying nonterminal CST node of kind `BitwiseAndExpression`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `BitwiseAndExpression`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `BitwiseAndExpression`.
     */
    cst: NonterminalNode);
    /**
     * Returns the child node that has the label `left_operand`.
     */
    get leftOperand(): Expression;
    /**
     * Returns the child node that has the label `operator`.
     */
    get operator(): TerminalNode;
    /**
     * Returns the child node that has the label `right_operand`.
     */
    get rightOperand(): Expression;
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
export declare class ShiftExpression {
    /**
     * The underlying nonterminal CST node of kind `ShiftExpression`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `ShiftExpression`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `ShiftExpression`.
     */
    cst: NonterminalNode);
    /**
     * Returns the child node that has the label `left_operand`.
     */
    get leftOperand(): Expression;
    /**
     * Returns the child node that has the label `operator`.
     */
    get operator(): TerminalNode;
    /**
     * Returns the child node that has the label `right_operand`.
     */
    get rightOperand(): Expression;
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
export declare class AdditiveExpression {
    /**
     * The underlying nonterminal CST node of kind `AdditiveExpression`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `AdditiveExpression`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `AdditiveExpression`.
     */
    cst: NonterminalNode);
    /**
     * Returns the child node that has the label `left_operand`.
     */
    get leftOperand(): Expression;
    /**
     * Returns the child node that has the label `operator`.
     */
    get operator(): TerminalNode;
    /**
     * Returns the child node that has the label `right_operand`.
     */
    get rightOperand(): Expression;
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
export declare class MultiplicativeExpression {
    /**
     * The underlying nonterminal CST node of kind `MultiplicativeExpression`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `MultiplicativeExpression`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `MultiplicativeExpression`.
     */
    cst: NonterminalNode);
    /**
     * Returns the child node that has the label `left_operand`.
     */
    get leftOperand(): Expression;
    /**
     * Returns the child node that has the label `operator`.
     */
    get operator(): TerminalNode;
    /**
     * Returns the child node that has the label `right_operand`.
     */
    get rightOperand(): Expression;
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
export declare class ExponentiationExpression {
    /**
     * The underlying nonterminal CST node of kind `ExponentiationExpression`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `ExponentiationExpression`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `ExponentiationExpression`.
     */
    cst: NonterminalNode);
    /**
     * Returns the child node that has the label `left_operand`.
     */
    get leftOperand(): Expression;
    /**
     * Returns the child node that has the label `operator`.
     */
    get operator(): TerminalNode;
    /**
     * Returns the child node that has the label `right_operand`.
     */
    get rightOperand(): Expression;
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
export declare class PostfixExpression {
    /**
     * The underlying nonterminal CST node of kind `PostfixExpression`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `PostfixExpression`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `PostfixExpression`.
     */
    cst: NonterminalNode);
    /**
     * Returns the child node that has the label `operand`.
     */
    get operand(): Expression;
    /**
     * Returns the child node that has the label `operator`.
     */
    get operator(): TerminalNode;
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
export declare class PrefixExpression {
    /**
     * The underlying nonterminal CST node of kind `PrefixExpression`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `PrefixExpression`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `PrefixExpression`.
     */
    cst: NonterminalNode);
    /**
     * Returns the child node that has the label `operator`.
     */
    get operator(): TerminalNode;
    /**
     * Returns the child node that has the label `operand`.
     */
    get operand(): Expression;
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
export declare class FunctionCallExpression {
    /**
     * The underlying nonterminal CST node of kind `FunctionCallExpression`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `FunctionCallExpression`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `FunctionCallExpression`.
     */
    cst: NonterminalNode);
    /**
     * Returns the child node that has the label `operand`.
     */
    get operand(): Expression;
    /**
     * Returns the child node that has the label `arguments`.
     */
    get arguments(): ArgumentsDeclaration;
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
export declare class CallOptionsExpression {
    /**
     * The underlying nonterminal CST node of kind `CallOptionsExpression`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `CallOptionsExpression`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `CallOptionsExpression`.
     */
    cst: NonterminalNode);
    /**
     * Returns the child node that has the label `operand`.
     */
    get operand(): Expression;
    /**
     * Returns the child node that has the label `open_brace`.
     */
    get openBrace(): TerminalNode;
    /**
     * Returns the child node that has the label `options`.
     */
    get options(): CallOptions;
    /**
     * Returns the child node that has the label `close_brace`.
     */
    get closeBrace(): TerminalNode;
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
export declare class MemberAccessExpression {
    /**
     * The underlying nonterminal CST node of kind `MemberAccessExpression`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `MemberAccessExpression`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `MemberAccessExpression`.
     */
    cst: NonterminalNode);
    /**
     * Returns the child node that has the label `operand`.
     */
    get operand(): Expression;
    /**
     * Returns the child node that has the label `period`.
     */
    get period(): TerminalNode;
    /**
     * Returns the child node that has the label `member`.
     */
    get member(): TerminalNode;
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
export declare class IndexAccessExpression {
    /**
     * The underlying nonterminal CST node of kind `IndexAccessExpression`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `IndexAccessExpression`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `IndexAccessExpression`.
     */
    cst: NonterminalNode);
    /**
     * Returns the child node that has the label `operand`.
     */
    get operand(): Expression;
    /**
     * Returns the child node that has the label `open_bracket`.
     */
    get openBracket(): TerminalNode;
    /**
     * Returns the child node that has the label `start`.
     */
    get start(): Expression | undefined;
    /**
     * Returns the child node that has the label `end`.
     */
    get end(): IndexAccessEnd | undefined;
    /**
     * Returns the child node that has the label `close_bracket`.
     */
    get closeBracket(): TerminalNode;
}
/**
 * This node represents a `IndexAccessEnd` nonterminal, with the following structure:
 *
 * ```ebnf
 * IndexAccessEnd = (* colon: *) COLON
 *                  (* end: *) Expression?;
 * ```
 */
export declare class IndexAccessEnd {
    /**
     * The underlying nonterminal CST node of kind `IndexAccessEnd`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `IndexAccessEnd`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `IndexAccessEnd`.
     */
    cst: NonterminalNode);
    /**
     * Returns the child node that has the label `colon`.
     */
    get colon(): TerminalNode;
    /**
     * Returns the child node that has the label `end`.
     */
    get end(): Expression | undefined;
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
export declare class PositionalArgumentsDeclaration {
    /**
     * The underlying nonterminal CST node of kind `PositionalArgumentsDeclaration`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `PositionalArgumentsDeclaration`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `PositionalArgumentsDeclaration`.
     */
    cst: NonterminalNode);
    /**
     * Returns the child node that has the label `open_paren`.
     */
    get openParen(): TerminalNode;
    /**
     * Returns the child node that has the label `arguments`.
     */
    get arguments(): PositionalArguments;
    /**
     * Returns the child node that has the label `close_paren`.
     */
    get closeParen(): TerminalNode;
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
export declare class NamedArgumentsDeclaration {
    /**
     * The underlying nonterminal CST node of kind `NamedArgumentsDeclaration`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `NamedArgumentsDeclaration`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `NamedArgumentsDeclaration`.
     */
    cst: NonterminalNode);
    /**
     * Returns the child node that has the label `open_paren`.
     */
    get openParen(): TerminalNode;
    /**
     * Returns the child node that has the label `arguments`.
     */
    get arguments(): NamedArgumentGroup | undefined;
    /**
     * Returns the child node that has the label `close_paren`.
     */
    get closeParen(): TerminalNode;
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
export declare class NamedArgumentGroup {
    /**
     * The underlying nonterminal CST node of kind `NamedArgumentGroup`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `NamedArgumentGroup`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `NamedArgumentGroup`.
     */
    cst: NonterminalNode);
    /**
     * Returns the child node that has the label `open_brace`.
     */
    get openBrace(): TerminalNode;
    /**
     * Returns the child node that has the label `arguments`.
     */
    get arguments(): NamedArguments;
    /**
     * Returns the child node that has the label `close_brace`.
     */
    get closeBrace(): TerminalNode;
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
export declare class NamedArgument {
    /**
     * The underlying nonterminal CST node of kind `NamedArgument`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `NamedArgument`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `NamedArgument`.
     */
    cst: NonterminalNode);
    /**
     * Returns the child node that has the label `name`.
     */
    get name(): TerminalNode;
    /**
     * Returns the child node that has the label `colon`.
     */
    get colon(): TerminalNode;
    /**
     * Returns the child node that has the label `value`.
     */
    get value(): Expression;
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
export declare class TypeExpression {
    /**
     * The underlying nonterminal CST node of kind `TypeExpression`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `TypeExpression`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `TypeExpression`.
     */
    cst: NonterminalNode);
    /**
     * Returns the child node that has the label `type_keyword`.
     */
    get typeKeyword(): TerminalNode;
    /**
     * Returns the child node that has the label `open_paren`.
     */
    get openParen(): TerminalNode;
    /**
     * Returns the child node that has the label `type_name`.
     */
    get typeName(): TypeName;
    /**
     * Returns the child node that has the label `close_paren`.
     */
    get closeParen(): TerminalNode;
}
/**
 * This node represents a `NewExpression` nonterminal, with the following structure:
 *
 * ```ebnf
 * NewExpression = (* new_keyword: *) NEW_KEYWORD
 *                 (* type_name: *) TypeName;
 * ```
 */
export declare class NewExpression {
    /**
     * The underlying nonterminal CST node of kind `NewExpression`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `NewExpression`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `NewExpression`.
     */
    cst: NonterminalNode);
    /**
     * Returns the child node that has the label `new_keyword`.
     */
    get newKeyword(): TerminalNode;
    /**
     * Returns the child node that has the label `type_name`.
     */
    get typeName(): TypeName;
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
export declare class TupleExpression {
    /**
     * The underlying nonterminal CST node of kind `TupleExpression`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `TupleExpression`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `TupleExpression`.
     */
    cst: NonterminalNode);
    /**
     * Returns the child node that has the label `open_paren`.
     */
    get openParen(): TerminalNode;
    /**
     * Returns the child node that has the label `items`.
     */
    get items(): TupleValues;
    /**
     * Returns the child node that has the label `close_paren`.
     */
    get closeParen(): TerminalNode;
}
/**
 * This node represents a `TupleValue` nonterminal, with the following structure:
 *
 * ```ebnf
 * TupleValue = (* expression: *) Expression?;
 * ```
 */
export declare class TupleValue {
    /**
     * The underlying nonterminal CST node of kind `TupleValue`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `TupleValue`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `TupleValue`.
     */
    cst: NonterminalNode);
    /**
     * Returns the child node that has the label `expression`.
     */
    get expression(): Expression | undefined;
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
export declare class ArrayExpression {
    /**
     * The underlying nonterminal CST node of kind `ArrayExpression`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `ArrayExpression`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `ArrayExpression`.
     */
    cst: NonterminalNode);
    /**
     * Returns the child node that has the label `open_bracket`.
     */
    get openBracket(): TerminalNode;
    /**
     * Returns the child node that has the label `items`.
     */
    get items(): ArrayValues;
    /**
     * Returns the child node that has the label `close_bracket`.
     */
    get closeBracket(): TerminalNode;
}
/**
 * This node represents a `HexNumberExpression` nonterminal, with the following structure:
 *
 * ```ebnf
 * HexNumberExpression = (* literal: *) HEX_LITERAL
 *                       (* unit: *) NumberUnit?; (* Deprecated in 0.5.0 *)
 * ```
 */
export declare class HexNumberExpression {
    /**
     * The underlying nonterminal CST node of kind `HexNumberExpression`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `HexNumberExpression`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `HexNumberExpression`.
     */
    cst: NonterminalNode);
    /**
     * Returns the child node that has the label `literal`.
     */
    get literal(): TerminalNode;
    /**
     * Returns the child node that has the label `unit`.
     */
    get unit(): NumberUnit | undefined;
}
/**
 * This node represents a `DecimalNumberExpression` nonterminal, with the following structure:
 *
 * ```ebnf
 * DecimalNumberExpression = (* literal: *) DECIMAL_LITERAL
 *                           (* unit: *) NumberUnit?;
 * ```
 */
export declare class DecimalNumberExpression {
    /**
     * The underlying nonterminal CST node of kind `DecimalNumberExpression`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `DecimalNumberExpression`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `DecimalNumberExpression`.
     */
    cst: NonterminalNode);
    /**
     * Returns the child node that has the label `literal`.
     */
    get literal(): TerminalNode;
    /**
     * Returns the child node that has the label `unit`.
     */
    get unit(): NumberUnit | undefined;
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
export declare class YulBlock {
    /**
     * The underlying nonterminal CST node of kind `YulBlock`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `YulBlock`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `YulBlock`.
     */
    cst: NonterminalNode);
    /**
     * Returns the child node that has the label `open_brace`.
     */
    get openBrace(): TerminalNode;
    /**
     * Returns the child node that has the label `statements`.
     */
    get statements(): YulStatements;
    /**
     * Returns the child node that has the label `close_brace`.
     */
    get closeBrace(): TerminalNode;
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
export declare class YulFunctionDefinition {
    /**
     * The underlying nonterminal CST node of kind `YulFunctionDefinition`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `YulFunctionDefinition`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `YulFunctionDefinition`.
     */
    cst: NonterminalNode);
    /**
     * Returns the child node that has the label `function_keyword`.
     */
    get functionKeyword(): TerminalNode;
    /**
     * Returns the child node that has the label `name`.
     */
    get name(): TerminalNode;
    /**
     * Returns the child node that has the label `parameters`.
     */
    get parameters(): YulParametersDeclaration;
    /**
     * Returns the child node that has the label `returns`.
     */
    get returns(): YulReturnsDeclaration | undefined;
    /**
     * Returns the child node that has the label `body`.
     */
    get body(): YulBlock;
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
export declare class YulParametersDeclaration {
    /**
     * The underlying nonterminal CST node of kind `YulParametersDeclaration`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `YulParametersDeclaration`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `YulParametersDeclaration`.
     */
    cst: NonterminalNode);
    /**
     * Returns the child node that has the label `open_paren`.
     */
    get openParen(): TerminalNode;
    /**
     * Returns the child node that has the label `parameters`.
     */
    get parameters(): YulParameters;
    /**
     * Returns the child node that has the label `close_paren`.
     */
    get closeParen(): TerminalNode;
}
/**
 * This node represents a `YulReturnsDeclaration` nonterminal, with the following structure:
 *
 * ```ebnf
 * YulReturnsDeclaration = (* minus_greater_than: *) MINUS_GREATER_THAN
 *                         (* variables: *) YulVariableNames;
 * ```
 */
export declare class YulReturnsDeclaration {
    /**
     * The underlying nonterminal CST node of kind `YulReturnsDeclaration`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `YulReturnsDeclaration`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `YulReturnsDeclaration`.
     */
    cst: NonterminalNode);
    /**
     * Returns the child node that has the label `minus_greater_than`.
     */
    get minusGreaterThan(): TerminalNode;
    /**
     * Returns the child node that has the label `variables`.
     */
    get variables(): YulVariableNames;
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
export declare class YulVariableDeclarationStatement {
    /**
     * The underlying nonterminal CST node of kind `YulVariableDeclarationStatement`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `YulVariableDeclarationStatement`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `YulVariableDeclarationStatement`.
     */
    cst: NonterminalNode);
    /**
     * Returns the child node that has the label `let_keyword`.
     */
    get letKeyword(): TerminalNode;
    /**
     * Returns the child node that has the label `variables`.
     */
    get variables(): YulVariableNames;
    /**
     * Returns the child node that has the label `value`.
     */
    get value(): YulVariableDeclarationValue | undefined;
}
/**
 * This node represents a `YulVariableDeclarationValue` nonterminal, with the following structure:
 *
 * ```ebnf
 * YulVariableDeclarationValue = (* assignment: *) YulAssignmentOperator
 *                               (* expression: *) YulExpression;
 * ```
 */
export declare class YulVariableDeclarationValue {
    /**
     * The underlying nonterminal CST node of kind `YulVariableDeclarationValue`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `YulVariableDeclarationValue`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `YulVariableDeclarationValue`.
     */
    cst: NonterminalNode);
    /**
     * Returns the child node that has the label `assignment`.
     */
    get assignment(): YulAssignmentOperator;
    /**
     * Returns the child node that has the label `expression`.
     */
    get expression(): YulExpression;
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
export declare class YulVariableAssignmentStatement {
    /**
     * The underlying nonterminal CST node of kind `YulVariableAssignmentStatement`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `YulVariableAssignmentStatement`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `YulVariableAssignmentStatement`.
     */
    cst: NonterminalNode);
    /**
     * Returns the child node that has the label `variables`.
     */
    get variables(): YulPaths;
    /**
     * Returns the child node that has the label `assignment`.
     */
    get assignment(): YulAssignmentOperator;
    /**
     * Returns the child node that has the label `expression`.
     */
    get expression(): YulExpression;
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
export declare class YulColonAndEqual {
    /**
     * The underlying nonterminal CST node of kind `YulColonAndEqual`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `YulColonAndEqual`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `YulColonAndEqual`.
     */
    cst: NonterminalNode);
    /**
     * Returns the child node that has the label `colon`.
     */
    get colon(): TerminalNode;
    /**
     * Returns the child node that has the label `equal`.
     */
    get equal(): TerminalNode;
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
export declare class YulStackAssignmentStatement {
    /**
     * The underlying nonterminal CST node of kind `YulStackAssignmentStatement`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `YulStackAssignmentStatement`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `YulStackAssignmentStatement`.
     */
    cst: NonterminalNode);
    /**
     * Returns the child node that has the label `assignment`.
     */
    get assignment(): YulStackAssignmentOperator;
    /**
     * Returns the child node that has the label `variable`.
     */
    get variable(): TerminalNode;
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
export declare class YulEqualAndColon {
    /**
     * The underlying nonterminal CST node of kind `YulEqualAndColon`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `YulEqualAndColon`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `YulEqualAndColon`.
     */
    cst: NonterminalNode);
    /**
     * Returns the child node that has the label `equal`.
     */
    get equal(): TerminalNode;
    /**
     * Returns the child node that has the label `colon`.
     */
    get colon(): TerminalNode;
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
export declare class YulIfStatement {
    /**
     * The underlying nonterminal CST node of kind `YulIfStatement`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `YulIfStatement`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `YulIfStatement`.
     */
    cst: NonterminalNode);
    /**
     * Returns the child node that has the label `if_keyword`.
     */
    get ifKeyword(): TerminalNode;
    /**
     * Returns the child node that has the label `condition`.
     */
    get condition(): YulExpression;
    /**
     * Returns the child node that has the label `body`.
     */
    get body(): YulBlock;
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
export declare class YulForStatement {
    /**
     * The underlying nonterminal CST node of kind `YulForStatement`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `YulForStatement`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `YulForStatement`.
     */
    cst: NonterminalNode);
    /**
     * Returns the child node that has the label `for_keyword`.
     */
    get forKeyword(): TerminalNode;
    /**
     * Returns the child node that has the label `initialization`.
     */
    get initialization(): YulBlock;
    /**
     * Returns the child node that has the label `condition`.
     */
    get condition(): YulExpression;
    /**
     * Returns the child node that has the label `iterator`.
     */
    get iterator(): YulBlock;
    /**
     * Returns the child node that has the label `body`.
     */
    get body(): YulBlock;
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
export declare class YulSwitchStatement {
    /**
     * The underlying nonterminal CST node of kind `YulSwitchStatement`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `YulSwitchStatement`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `YulSwitchStatement`.
     */
    cst: NonterminalNode);
    /**
     * Returns the child node that has the label `switch_keyword`.
     */
    get switchKeyword(): TerminalNode;
    /**
     * Returns the child node that has the label `expression`.
     */
    get expression(): YulExpression;
    /**
     * Returns the child node that has the label `cases`.
     */
    get cases(): YulSwitchCases;
}
/**
 * This node represents a `YulDefaultCase` nonterminal, with the following structure:
 *
 * ```ebnf
 * YulDefaultCase = (* default_keyword: *) YUL_DEFAULT_KEYWORD
 *                  (* body: *) YulBlock;
 * ```
 */
export declare class YulDefaultCase {
    /**
     * The underlying nonterminal CST node of kind `YulDefaultCase`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `YulDefaultCase`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `YulDefaultCase`.
     */
    cst: NonterminalNode);
    /**
     * Returns the child node that has the label `default_keyword`.
     */
    get defaultKeyword(): TerminalNode;
    /**
     * Returns the child node that has the label `body`.
     */
    get body(): YulBlock;
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
export declare class YulValueCase {
    /**
     * The underlying nonterminal CST node of kind `YulValueCase`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `YulValueCase`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `YulValueCase`.
     */
    cst: NonterminalNode);
    /**
     * Returns the child node that has the label `case_keyword`.
     */
    get caseKeyword(): TerminalNode;
    /**
     * Returns the child node that has the label `value`.
     */
    get value(): YulLiteral;
    /**
     * Returns the child node that has the label `body`.
     */
    get body(): YulBlock;
}
/**
 * This node represents a `YulLeaveStatement` nonterminal, with the following structure:
 *
 * ```ebnf
 * (* Introduced in 0.6.0 *)
 * YulLeaveStatement = (* leave_keyword: *) YUL_LEAVE_KEYWORD;
 * ```
 */
export declare class YulLeaveStatement {
    /**
     * The underlying nonterminal CST node of kind `YulLeaveStatement`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `YulLeaveStatement`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `YulLeaveStatement`.
     */
    cst: NonterminalNode);
    /**
     * Returns the child node that has the label `leave_keyword`.
     */
    get leaveKeyword(): TerminalNode;
}
/**
 * This node represents a `YulBreakStatement` nonterminal, with the following structure:
 *
 * ```ebnf
 * YulBreakStatement = (* break_keyword: *) YUL_BREAK_KEYWORD;
 * ```
 */
export declare class YulBreakStatement {
    /**
     * The underlying nonterminal CST node of kind `YulBreakStatement`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `YulBreakStatement`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `YulBreakStatement`.
     */
    cst: NonterminalNode);
    /**
     * Returns the child node that has the label `break_keyword`.
     */
    get breakKeyword(): TerminalNode;
}
/**
 * This node represents a `YulContinueStatement` nonterminal, with the following structure:
 *
 * ```ebnf
 * YulContinueStatement = (* continue_keyword: *) YUL_CONTINUE_KEYWORD;
 * ```
 */
export declare class YulContinueStatement {
    /**
     * The underlying nonterminal CST node of kind `YulContinueStatement`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `YulContinueStatement`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `YulContinueStatement`.
     */
    cst: NonterminalNode);
    /**
     * Returns the child node that has the label `continue_keyword`.
     */
    get continueKeyword(): TerminalNode;
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
export declare class YulLabel {
    /**
     * The underlying nonterminal CST node of kind `YulLabel`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `YulLabel`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `YulLabel`.
     */
    cst: NonterminalNode);
    /**
     * Returns the child node that has the label `label`.
     */
    get label(): TerminalNode;
    /**
     * Returns the child node that has the label `colon`.
     */
    get colon(): TerminalNode;
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
export declare class YulFunctionCallExpression {
    /**
     * The underlying nonterminal CST node of kind `YulFunctionCallExpression`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `YulFunctionCallExpression`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `YulFunctionCallExpression`.
     */
    cst: NonterminalNode);
    /**
     * Returns the child node that has the label `operand`.
     */
    get operand(): YulExpression;
    /**
     * Returns the child node that has the label `open_paren`.
     */
    get openParen(): TerminalNode;
    /**
     * Returns the child node that has the label `arguments`.
     */
    get arguments(): YulArguments;
    /**
     * Returns the child node that has the label `close_paren`.
     */
    get closeParen(): TerminalNode;
}
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
export declare class SourceUnitMember {
    /**
     * The underlying nonterminal CST node of kind `SourceUnitMember`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `SourceUnitMember`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `SourceUnitMember`.
     */
    cst: NonterminalNode);
    /**
     * Returns the child node that has the label `variant`.
     */
    get variant(): PragmaDirective | ImportDirective | ContractDefinition | InterfaceDefinition | LibraryDefinition | StructDefinition | EnumDefinition | FunctionDefinition | ErrorDefinition | UserDefinedValueTypeDefinition | UsingDirective | EventDefinition | ConstantDefinition;
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
export declare class Pragma {
    /**
     * The underlying nonterminal CST node of kind `Pragma`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `Pragma`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `Pragma`.
     */
    cst: NonterminalNode);
    /**
     * Returns the child node that has the label `variant`.
     */
    get variant(): VersionPragma | AbicoderPragma | ExperimentalPragma;
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
export declare class AbicoderVersion {
    /**
     * The underlying nonterminal CST node of kind `AbicoderVersion`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `AbicoderVersion`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `AbicoderVersion`.
     */
    cst: NonterminalNode);
    /**
     * Returns the child node that has the label `variant`.
     */
    get variant(): TerminalNode;
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
export declare class ExperimentalFeature {
    /**
     * The underlying nonterminal CST node of kind `ExperimentalFeature`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `ExperimentalFeature`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `ExperimentalFeature`.
     */
    cst: NonterminalNode);
    /**
     * Returns the child node that has the label `variant`.
     */
    get variant(): StringLiteral | TerminalNode;
}
/**
 * This node represents a `VersionExpression` nonterminal, with the following structure:
 *
 * ```ebnf
 * VersionExpression = (* variant: *) VersionRange
 *                   | (* variant: *) VersionTerm;
 * ```
 */
export declare class VersionExpression {
    /**
     * The underlying nonterminal CST node of kind `VersionExpression`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `VersionExpression`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `VersionExpression`.
     */
    cst: NonterminalNode);
    /**
     * Returns the child node that has the label `variant`.
     */
    get variant(): VersionRange | VersionTerm;
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
export declare class VersionOperator {
    /**
     * The underlying nonterminal CST node of kind `VersionOperator`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `VersionOperator`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `VersionOperator`.
     */
    cst: NonterminalNode);
    /**
     * Returns the child node that has the label `variant`.
     */
    get variant(): TerminalNode;
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
export declare class VersionLiteral {
    /**
     * The underlying nonterminal CST node of kind `VersionLiteral`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `VersionLiteral`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `VersionLiteral`.
     */
    cst: NonterminalNode);
    /**
     * Returns the child node that has the label `variant`.
     */
    get variant(): SimpleVersionLiteral | TerminalNode;
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
export declare class ImportClause {
    /**
     * The underlying nonterminal CST node of kind `ImportClause`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `ImportClause`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `ImportClause`.
     */
    cst: NonterminalNode);
    /**
     * Returns the child node that has the label `variant`.
     */
    get variant(): PathImport | NamedImport | ImportDeconstruction;
}
/**
 * This node represents a `UsingClause` nonterminal, with the following structure:
 *
 * ```ebnf
 * UsingClause = (* variant: *) IdentifierPath
 *             | (* variant: *) UsingDeconstruction; (* Introduced in 0.8.13 *)
 * ```
 */
export declare class UsingClause {
    /**
     * The underlying nonterminal CST node of kind `UsingClause`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `UsingClause`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `UsingClause`.
     */
    cst: NonterminalNode);
    /**
     * Returns the child node that has the label `variant`.
     */
    get variant(): IdentifierPath | UsingDeconstruction;
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
export declare class UsingOperator {
    /**
     * The underlying nonterminal CST node of kind `UsingOperator`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `UsingOperator`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `UsingOperator`.
     */
    cst: NonterminalNode);
    /**
     * Returns the child node that has the label `variant`.
     */
    get variant(): TerminalNode;
}
/**
 * This node represents a `UsingTarget` nonterminal, with the following structure:
 *
 * ```ebnf
 * UsingTarget = (* variant: *) TypeName
 *             | (* variant: *) ASTERISK;
 * ```
 */
export declare class UsingTarget {
    /**
     * The underlying nonterminal CST node of kind `UsingTarget`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `UsingTarget`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `UsingTarget`.
     */
    cst: NonterminalNode);
    /**
     * Returns the child node that has the label `variant`.
     */
    get variant(): TypeName | TerminalNode;
}
/**
 * This node represents a `ContractSpecifier` nonterminal, with the following structure:
 *
 * ```ebnf
 * ContractSpecifier = (* variant: *) InheritanceSpecifier
 *                   | (* variant: *) StorageLayoutSpecifier; (* Introduced in 0.8.29 *)
 * ```
 */
export declare class ContractSpecifier {
    /**
     * The underlying nonterminal CST node of kind `ContractSpecifier`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `ContractSpecifier`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `ContractSpecifier`.
     */
    cst: NonterminalNode);
    /**
     * Returns the child node that has the label `variant`.
     */
    get variant(): InheritanceSpecifier | StorageLayoutSpecifier;
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
export declare class ContractMember {
    /**
     * The underlying nonterminal CST node of kind `ContractMember`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `ContractMember`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `ContractMember`.
     */
    cst: NonterminalNode);
    /**
     * Returns the child node that has the label `variant`.
     */
    get variant(): UsingDirective | FunctionDefinition | ConstructorDefinition | ReceiveFunctionDefinition | FallbackFunctionDefinition | UnnamedFunctionDefinition | ModifierDefinition | StructDefinition | EnumDefinition | EventDefinition | ErrorDefinition | UserDefinedValueTypeDefinition | StateVariableDefinition;
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
export declare class StateVariableAttribute {
    /**
     * The underlying nonterminal CST node of kind `StateVariableAttribute`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `StateVariableAttribute`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `StateVariableAttribute`.
     */
    cst: NonterminalNode);
    /**
     * Returns the child node that has the label `variant`.
     */
    get variant(): OverrideSpecifier | TerminalNode;
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
export declare class FunctionName {
    /**
     * The underlying nonterminal CST node of kind `FunctionName`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `FunctionName`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `FunctionName`.
     */
    cst: NonterminalNode);
    /**
     * Returns the child node that has the label `variant`.
     */
    get variant(): TerminalNode;
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
export declare class FunctionAttribute {
    /**
     * The underlying nonterminal CST node of kind `FunctionAttribute`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `FunctionAttribute`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `FunctionAttribute`.
     */
    cst: NonterminalNode);
    /**
     * Returns the child node that has the label `variant`.
     */
    get variant(): ModifierInvocation | OverrideSpecifier | TerminalNode;
}
/**
 * This node represents a `FunctionBody` nonterminal, with the following structure:
 *
 * ```ebnf
 * FunctionBody = (* variant: *) Block
 *              | (* variant: *) SEMICOLON;
 * ```
 */
export declare class FunctionBody {
    /**
     * The underlying nonterminal CST node of kind `FunctionBody`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `FunctionBody`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `FunctionBody`.
     */
    cst: NonterminalNode);
    /**
     * Returns the child node that has the label `variant`.
     */
    get variant(): Block | TerminalNode;
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
export declare class ConstructorAttribute {
    /**
     * The underlying nonterminal CST node of kind `ConstructorAttribute`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `ConstructorAttribute`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `ConstructorAttribute`.
     */
    cst: NonterminalNode);
    /**
     * Returns the child node that has the label `variant`.
     */
    get variant(): ModifierInvocation | TerminalNode;
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
export declare class UnnamedFunctionAttribute {
    /**
     * The underlying nonterminal CST node of kind `UnnamedFunctionAttribute`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `UnnamedFunctionAttribute`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `UnnamedFunctionAttribute`.
     */
    cst: NonterminalNode);
    /**
     * Returns the child node that has the label `variant`.
     */
    get variant(): ModifierInvocation | TerminalNode;
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
export declare class FallbackFunctionAttribute {
    /**
     * The underlying nonterminal CST node of kind `FallbackFunctionAttribute`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `FallbackFunctionAttribute`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `FallbackFunctionAttribute`.
     */
    cst: NonterminalNode);
    /**
     * Returns the child node that has the label `variant`.
     */
    get variant(): ModifierInvocation | OverrideSpecifier | TerminalNode;
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
export declare class ReceiveFunctionAttribute {
    /**
     * The underlying nonterminal CST node of kind `ReceiveFunctionAttribute`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `ReceiveFunctionAttribute`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `ReceiveFunctionAttribute`.
     */
    cst: NonterminalNode);
    /**
     * Returns the child node that has the label `variant`.
     */
    get variant(): ModifierInvocation | OverrideSpecifier | TerminalNode;
}
/**
 * This node represents a `ModifierAttribute` nonterminal, with the following structure:
 *
 * ```ebnf
 * ModifierAttribute = (* variant: *) OverrideSpecifier (* Introduced in 0.6.0 *)
 *                   | (* variant: *) VIRTUAL_KEYWORD; (* Introduced in 0.6.0 *)
 * ```
 */
export declare class ModifierAttribute {
    /**
     * The underlying nonterminal CST node of kind `ModifierAttribute`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `ModifierAttribute`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `ModifierAttribute`.
     */
    cst: NonterminalNode);
    /**
     * Returns the child node that has the label `variant`.
     */
    get variant(): OverrideSpecifier | TerminalNode;
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
export declare class TypeName {
    /**
     * The underlying nonterminal CST node of kind `TypeName`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `TypeName`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `TypeName`.
     */
    cst: NonterminalNode);
    /**
     * Returns the child node that has the label `variant`.
     */
    get variant(): ArrayTypeName | FunctionType | MappingType | ElementaryType | IdentifierPath;
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
export declare class FunctionTypeAttribute {
    /**
     * The underlying nonterminal CST node of kind `FunctionTypeAttribute`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `FunctionTypeAttribute`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `FunctionTypeAttribute`.
     */
    cst: NonterminalNode);
    /**
     * Returns the child node that has the label `variant`.
     */
    get variant(): TerminalNode;
}
/**
 * This node represents a `MappingKeyType` nonterminal, with the following structure:
 *
 * ```ebnf
 * MappingKeyType = (* variant: *) ElementaryType
 *                | (* variant: *) IdentifierPath;
 * ```
 */
export declare class MappingKeyType {
    /**
     * The underlying nonterminal CST node of kind `MappingKeyType`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `MappingKeyType`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `MappingKeyType`.
     */
    cst: NonterminalNode);
    /**
     * Returns the child node that has the label `variant`.
     */
    get variant(): ElementaryType | IdentifierPath;
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
export declare class ElementaryType {
    /**
     * The underlying nonterminal CST node of kind `ElementaryType`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `ElementaryType`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `ElementaryType`.
     */
    cst: NonterminalNode);
    /**
     * Returns the child node that has the label `variant`.
     */
    get variant(): AddressType | TerminalNode;
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
export declare class Statement {
    /**
     * The underlying nonterminal CST node of kind `Statement`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `Statement`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `Statement`.
     */
    cst: NonterminalNode);
    /**
     * Returns the child node that has the label `variant`.
     */
    get variant(): IfStatement | ForStatement | WhileStatement | DoWhileStatement | ContinueStatement | BreakStatement | ReturnStatement | ThrowStatement | EmitStatement | TryStatement | RevertStatement | AssemblyStatement | Block | UncheckedBlock | TupleDeconstructionStatement | VariableDeclarationStatement | ExpressionStatement;
}
/**
 * This node represents a `TupleMember` nonterminal, with the following structure:
 *
 * ```ebnf
 * TupleMember = (* variant: *) TypedTupleMember
 *             | (* variant: *) UntypedTupleMember;
 * ```
 */
export declare class TupleMember {
    /**
     * The underlying nonterminal CST node of kind `TupleMember`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `TupleMember`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `TupleMember`.
     */
    cst: NonterminalNode);
    /**
     * Returns the child node that has the label `variant`.
     */
    get variant(): TypedTupleMember | UntypedTupleMember;
}
/**
 * This node represents a `VariableDeclarationType` nonterminal, with the following structure:
 *
 * ```ebnf
 * VariableDeclarationType = (* variant: *) TypeName
 *                         | (* variant: *) VAR_KEYWORD; (* Deprecated in 0.5.0 *)
 * ```
 */
export declare class VariableDeclarationType {
    /**
     * The underlying nonterminal CST node of kind `VariableDeclarationType`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `VariableDeclarationType`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `VariableDeclarationType`.
     */
    cst: NonterminalNode);
    /**
     * Returns the child node that has the label `variant`.
     */
    get variant(): TypeName | TerminalNode;
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
export declare class StorageLocation {
    /**
     * The underlying nonterminal CST node of kind `StorageLocation`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `StorageLocation`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `StorageLocation`.
     */
    cst: NonterminalNode);
    /**
     * Returns the child node that has the label `variant`.
     */
    get variant(): TerminalNode;
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
export declare class ForStatementInitialization {
    /**
     * The underlying nonterminal CST node of kind `ForStatementInitialization`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `ForStatementInitialization`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `ForStatementInitialization`.
     */
    cst: NonterminalNode);
    /**
     * Returns the child node that has the label `variant`.
     */
    get variant(): TupleDeconstructionStatement | VariableDeclarationStatement | ExpressionStatement | TerminalNode;
}
/**
 * This node represents a `ForStatementCondition` nonterminal, with the following structure:
 *
 * ```ebnf
 * ForStatementCondition = (* variant: *) ExpressionStatement
 *                       | (* variant: *) SEMICOLON;
 * ```
 */
export declare class ForStatementCondition {
    /**
     * The underlying nonterminal CST node of kind `ForStatementCondition`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `ForStatementCondition`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `ForStatementCondition`.
     */
    cst: NonterminalNode);
    /**
     * Returns the child node that has the label `variant`.
     */
    get variant(): ExpressionStatement | TerminalNode;
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
export declare class Expression {
    /**
     * The underlying nonterminal CST node of kind `Expression`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `Expression`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `Expression`.
     */
    cst: NonterminalNode);
    /**
     * Returns the child node that has the label `variant`.
     */
    get variant(): AssignmentExpression | ConditionalExpression | OrExpression | AndExpression | EqualityExpression | InequalityExpression | BitwiseOrExpression | BitwiseXorExpression | BitwiseAndExpression | ShiftExpression | AdditiveExpression | MultiplicativeExpression | ExponentiationExpression | PostfixExpression | PrefixExpression | FunctionCallExpression | CallOptionsExpression | MemberAccessExpression | IndexAccessExpression | NewExpression | TupleExpression | TypeExpression | ArrayExpression | HexNumberExpression | DecimalNumberExpression | StringExpression | ElementaryType | TerminalNode;
}
/**
 * This node represents a `ArgumentsDeclaration` nonterminal, with the following structure:
 *
 * ```ebnf
 * ArgumentsDeclaration = (* variant: *) PositionalArgumentsDeclaration
 *                      | (* variant: *) NamedArgumentsDeclaration;
 * ```
 */
export declare class ArgumentsDeclaration {
    /**
     * The underlying nonterminal CST node of kind `ArgumentsDeclaration`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `ArgumentsDeclaration`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `ArgumentsDeclaration`.
     */
    cst: NonterminalNode);
    /**
     * Returns the child node that has the label `variant`.
     */
    get variant(): PositionalArgumentsDeclaration | NamedArgumentsDeclaration;
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
export declare class NumberUnit {
    /**
     * The underlying nonterminal CST node of kind `NumberUnit`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `NumberUnit`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `NumberUnit`.
     */
    cst: NonterminalNode);
    /**
     * Returns the child node that has the label `variant`.
     */
    get variant(): TerminalNode;
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
export declare class StringExpression {
    /**
     * The underlying nonterminal CST node of kind `StringExpression`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `StringExpression`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `StringExpression`.
     */
    cst: NonterminalNode);
    /**
     * Returns the child node that has the label `variant`.
     */
    get variant(): StringLiteral | StringLiterals | HexStringLiteral | HexStringLiterals | UnicodeStringLiterals;
}
/**
 * This node represents a `StringLiteral` nonterminal, with the following structure:
 *
 * ```ebnf
 * StringLiteral = (* variant: *) SINGLE_QUOTED_STRING_LITERAL
 *               | (* variant: *) DOUBLE_QUOTED_STRING_LITERAL;
 * ```
 */
export declare class StringLiteral {
    /**
     * The underlying nonterminal CST node of kind `StringLiteral`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `StringLiteral`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `StringLiteral`.
     */
    cst: NonterminalNode);
    /**
     * Returns the child node that has the label `variant`.
     */
    get variant(): TerminalNode;
}
/**
 * This node represents a `HexStringLiteral` nonterminal, with the following structure:
 *
 * ```ebnf
 * HexStringLiteral = (* variant: *) SINGLE_QUOTED_HEX_STRING_LITERAL
 *                  | (* variant: *) DOUBLE_QUOTED_HEX_STRING_LITERAL;
 * ```
 */
export declare class HexStringLiteral {
    /**
     * The underlying nonterminal CST node of kind `HexStringLiteral`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `HexStringLiteral`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `HexStringLiteral`.
     */
    cst: NonterminalNode);
    /**
     * Returns the child node that has the label `variant`.
     */
    get variant(): TerminalNode;
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
export declare class UnicodeStringLiteral {
    /**
     * The underlying nonterminal CST node of kind `UnicodeStringLiteral`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `UnicodeStringLiteral`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `UnicodeStringLiteral`.
     */
    cst: NonterminalNode);
    /**
     * Returns the child node that has the label `variant`.
     */
    get variant(): TerminalNode;
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
export declare class YulStatement {
    /**
     * The underlying nonterminal CST node of kind `YulStatement`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `YulStatement`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `YulStatement`.
     */
    cst: NonterminalNode);
    /**
     * Returns the child node that has the label `variant`.
     */
    get variant(): YulBlock | YulFunctionDefinition | YulStackAssignmentStatement | YulIfStatement | YulForStatement | YulSwitchStatement | YulLeaveStatement | YulBreakStatement | YulContinueStatement | YulVariableAssignmentStatement | YulLabel | YulVariableDeclarationStatement | YulExpression;
}
/**
 * This node represents a `YulAssignmentOperator` nonterminal, with the following structure:
 *
 * ```ebnf
 * YulAssignmentOperator = (* variant: *) COLON_EQUAL
 *                       | (* variant: *) YulColonAndEqual; (* Deprecated in 0.5.5 *)
 * ```
 */
export declare class YulAssignmentOperator {
    /**
     * The underlying nonterminal CST node of kind `YulAssignmentOperator`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `YulAssignmentOperator`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `YulAssignmentOperator`.
     */
    cst: NonterminalNode);
    /**
     * Returns the child node that has the label `variant`.
     */
    get variant(): YulColonAndEqual | TerminalNode;
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
export declare class YulStackAssignmentOperator {
    /**
     * The underlying nonterminal CST node of kind `YulStackAssignmentOperator`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `YulStackAssignmentOperator`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `YulStackAssignmentOperator`.
     */
    cst: NonterminalNode);
    /**
     * Returns the child node that has the label `variant`.
     */
    get variant(): YulEqualAndColon | TerminalNode;
}
/**
 * This node represents a `YulSwitchCase` nonterminal, with the following structure:
 *
 * ```ebnf
 * YulSwitchCase = (* variant: *) YulDefaultCase
 *               | (* variant: *) YulValueCase;
 * ```
 */
export declare class YulSwitchCase {
    /**
     * The underlying nonterminal CST node of kind `YulSwitchCase`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `YulSwitchCase`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `YulSwitchCase`.
     */
    cst: NonterminalNode);
    /**
     * Returns the child node that has the label `variant`.
     */
    get variant(): YulDefaultCase | YulValueCase;
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
export declare class YulExpression {
    /**
     * The underlying nonterminal CST node of kind `YulExpression`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `YulExpression`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `YulExpression`.
     */
    cst: NonterminalNode);
    /**
     * Returns the child node that has the label `variant`.
     */
    get variant(): YulFunctionCallExpression | YulLiteral | YulPath;
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
export declare class YulLiteral {
    /**
     * The underlying nonterminal CST node of kind `YulLiteral`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `YulLiteral`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `YulLiteral`.
     */
    cst: NonterminalNode);
    /**
     * Returns the child node that has the label `variant`.
     */
    get variant(): HexStringLiteral | StringLiteral | TerminalNode;
}
/**
 * This node represents a `SourceUnitMembers` nonterminal, with the following structure:
 *
 * ```ebnf
 * SourceUnitMembers = (* item: *) SourceUnitMember*;
 * ```
 */
export declare class SourceUnitMembers {
    /**
     * The underlying nonterminal CST node of kind `SourceUnitMembers`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `SourceUnitMembers`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `SourceUnitMembers`.
     */
    cst: NonterminalNode);
    /**
     * Returns an array of the child nodes that have the label `item`.
     */
    get items(): readonly SourceUnitMember[];
}
/**
 * This node represents a `VersionExpressionSet` nonterminal, with the following structure:
 *
 * ```ebnf
 * VersionExpressionSet = (* item: *) VersionExpression+;
 * ```
 */
export declare class VersionExpressionSet {
    /**
     * The underlying nonterminal CST node of kind `VersionExpressionSet`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `VersionExpressionSet`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `VersionExpressionSet`.
     */
    cst: NonterminalNode);
    /**
     * Returns an array of the child nodes that have the label `item`.
     */
    get items(): readonly VersionExpression[];
}
/**
 * This node represents a `ContractSpecifiers` nonterminal, with the following structure:
 *
 * ```ebnf
 * ContractSpecifiers = (* item: *) ContractSpecifier*;
 * ```
 */
export declare class ContractSpecifiers {
    /**
     * The underlying nonterminal CST node of kind `ContractSpecifiers`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `ContractSpecifiers`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `ContractSpecifiers`.
     */
    cst: NonterminalNode);
    /**
     * Returns an array of the child nodes that have the label `item`.
     */
    get items(): readonly ContractSpecifier[];
}
/**
 * This node represents a `ContractMembers` nonterminal, with the following structure:
 *
 * ```ebnf
 * ContractMembers = (* item: *) ContractMember*;
 * ```
 */
export declare class ContractMembers {
    /**
     * The underlying nonterminal CST node of kind `ContractMembers`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `ContractMembers`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `ContractMembers`.
     */
    cst: NonterminalNode);
    /**
     * Returns an array of the child nodes that have the label `item`.
     */
    get items(): readonly ContractMember[];
}
/**
 * This node represents a `InterfaceMembers` nonterminal, with the following structure:
 *
 * ```ebnf
 * InterfaceMembers = (* item: *) ContractMember*;
 * ```
 */
export declare class InterfaceMembers {
    /**
     * The underlying nonterminal CST node of kind `InterfaceMembers`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `InterfaceMembers`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `InterfaceMembers`.
     */
    cst: NonterminalNode);
    /**
     * Returns an array of the child nodes that have the label `item`.
     */
    get items(): readonly ContractMember[];
}
/**
 * This node represents a `LibraryMembers` nonterminal, with the following structure:
 *
 * ```ebnf
 * LibraryMembers = (* item: *) ContractMember*;
 * ```
 */
export declare class LibraryMembers {
    /**
     * The underlying nonterminal CST node of kind `LibraryMembers`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `LibraryMembers`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `LibraryMembers`.
     */
    cst: NonterminalNode);
    /**
     * Returns an array of the child nodes that have the label `item`.
     */
    get items(): readonly ContractMember[];
}
/**
 * This node represents a `StructMembers` nonterminal, with the following structure:
 *
 * ```ebnf
 * StructMembers = (* item: *) StructMember*;
 * ```
 */
export declare class StructMembers {
    /**
     * The underlying nonterminal CST node of kind `StructMembers`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `StructMembers`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `StructMembers`.
     */
    cst: NonterminalNode);
    /**
     * Returns an array of the child nodes that have the label `item`.
     */
    get items(): readonly StructMember[];
}
/**
 * This node represents a `StateVariableAttributes` nonterminal, with the following structure:
 *
 * ```ebnf
 * StateVariableAttributes = (* item: *) StateVariableAttribute*;
 * ```
 */
export declare class StateVariableAttributes {
    /**
     * The underlying nonterminal CST node of kind `StateVariableAttributes`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `StateVariableAttributes`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `StateVariableAttributes`.
     */
    cst: NonterminalNode);
    /**
     * Returns an array of the child nodes that have the label `item`.
     */
    get items(): readonly StateVariableAttribute[];
}
/**
 * This node represents a `FunctionAttributes` nonterminal, with the following structure:
 *
 * ```ebnf
 * FunctionAttributes = (* item: *) FunctionAttribute*;
 * ```
 */
export declare class FunctionAttributes {
    /**
     * The underlying nonterminal CST node of kind `FunctionAttributes`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `FunctionAttributes`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `FunctionAttributes`.
     */
    cst: NonterminalNode);
    /**
     * Returns an array of the child nodes that have the label `item`.
     */
    get items(): readonly FunctionAttribute[];
}
/**
 * This node represents a `ConstructorAttributes` nonterminal, with the following structure:
 *
 * ```ebnf
 * (* Introduced in 0.4.22 *)
 * ConstructorAttributes = (* item: *) ConstructorAttribute*;
 * ```
 */
export declare class ConstructorAttributes {
    /**
     * The underlying nonterminal CST node of kind `ConstructorAttributes`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `ConstructorAttributes`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `ConstructorAttributes`.
     */
    cst: NonterminalNode);
    /**
     * Returns an array of the child nodes that have the label `item`.
     */
    get items(): readonly ConstructorAttribute[];
}
/**
 * This node represents a `UnnamedFunctionAttributes` nonterminal, with the following structure:
 *
 * ```ebnf
 * (* Deprecated in 0.6.0 *)
 * UnnamedFunctionAttributes = (* item: *) UnnamedFunctionAttribute*;
 * ```
 */
export declare class UnnamedFunctionAttributes {
    /**
     * The underlying nonterminal CST node of kind `UnnamedFunctionAttributes`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `UnnamedFunctionAttributes`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `UnnamedFunctionAttributes`.
     */
    cst: NonterminalNode);
    /**
     * Returns an array of the child nodes that have the label `item`.
     */
    get items(): readonly UnnamedFunctionAttribute[];
}
/**
 * This node represents a `FallbackFunctionAttributes` nonterminal, with the following structure:
 *
 * ```ebnf
 * (* Introduced in 0.6.0 *)
 * FallbackFunctionAttributes = (* item: *) FallbackFunctionAttribute*;
 * ```
 */
export declare class FallbackFunctionAttributes {
    /**
     * The underlying nonterminal CST node of kind `FallbackFunctionAttributes`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `FallbackFunctionAttributes`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `FallbackFunctionAttributes`.
     */
    cst: NonterminalNode);
    /**
     * Returns an array of the child nodes that have the label `item`.
     */
    get items(): readonly FallbackFunctionAttribute[];
}
/**
 * This node represents a `ReceiveFunctionAttributes` nonterminal, with the following structure:
 *
 * ```ebnf
 * (* Introduced in 0.6.0 *)
 * ReceiveFunctionAttributes = (* item: *) ReceiveFunctionAttribute*;
 * ```
 */
export declare class ReceiveFunctionAttributes {
    /**
     * The underlying nonterminal CST node of kind `ReceiveFunctionAttributes`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `ReceiveFunctionAttributes`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `ReceiveFunctionAttributes`.
     */
    cst: NonterminalNode);
    /**
     * Returns an array of the child nodes that have the label `item`.
     */
    get items(): readonly ReceiveFunctionAttribute[];
}
/**
 * This node represents a `ModifierAttributes` nonterminal, with the following structure:
 *
 * ```ebnf
 * ModifierAttributes = (* item: *) ModifierAttribute*;
 * ```
 */
export declare class ModifierAttributes {
    /**
     * The underlying nonterminal CST node of kind `ModifierAttributes`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `ModifierAttributes`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `ModifierAttributes`.
     */
    cst: NonterminalNode);
    /**
     * Returns an array of the child nodes that have the label `item`.
     */
    get items(): readonly ModifierAttribute[];
}
/**
 * This node represents a `FunctionTypeAttributes` nonterminal, with the following structure:
 *
 * ```ebnf
 * FunctionTypeAttributes = (* item: *) FunctionTypeAttribute*;
 * ```
 */
export declare class FunctionTypeAttributes {
    /**
     * The underlying nonterminal CST node of kind `FunctionTypeAttributes`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `FunctionTypeAttributes`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `FunctionTypeAttributes`.
     */
    cst: NonterminalNode);
    /**
     * Returns an array of the child nodes that have the label `item`.
     */
    get items(): readonly FunctionTypeAttribute[];
}
/**
 * This node represents a `Statements` nonterminal, with the following structure:
 *
 * ```ebnf
 * Statements = (* item: *) Statement*;
 * ```
 */
export declare class Statements {
    /**
     * The underlying nonterminal CST node of kind `Statements`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `Statements`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `Statements`.
     */
    cst: NonterminalNode);
    /**
     * Returns an array of the child nodes that have the label `item`.
     */
    get items(): readonly Statement[];
}
/**
 * This node represents a `CatchClauses` nonterminal, with the following structure:
 *
 * ```ebnf
 * (* Introduced in 0.6.0 *)
 * CatchClauses = (* item: *) CatchClause+;
 * ```
 */
export declare class CatchClauses {
    /**
     * The underlying nonterminal CST node of kind `CatchClauses`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `CatchClauses`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `CatchClauses`.
     */
    cst: NonterminalNode);
    /**
     * Returns an array of the child nodes that have the label `item`.
     */
    get items(): readonly CatchClause[];
}
/**
 * This node represents a `StringLiterals` nonterminal, with the following structure:
 *
 * ```ebnf
 * (* Introduced in 0.5.14 *)
 * StringLiterals = (* item: *) StringLiteral+;
 * ```
 */
export declare class StringLiterals {
    /**
     * The underlying nonterminal CST node of kind `StringLiterals`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `StringLiterals`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `StringLiterals`.
     */
    cst: NonterminalNode);
    /**
     * Returns an array of the child nodes that have the label `item`.
     */
    get items(): readonly StringLiteral[];
}
/**
 * This node represents a `HexStringLiterals` nonterminal, with the following structure:
 *
 * ```ebnf
 * (* Introduced in 0.5.14 *)
 * HexStringLiterals = (* item: *) HexStringLiteral+;
 * ```
 */
export declare class HexStringLiterals {
    /**
     * The underlying nonterminal CST node of kind `HexStringLiterals`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `HexStringLiterals`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `HexStringLiterals`.
     */
    cst: NonterminalNode);
    /**
     * Returns an array of the child nodes that have the label `item`.
     */
    get items(): readonly HexStringLiteral[];
}
/**
 * This node represents a `UnicodeStringLiterals` nonterminal, with the following structure:
 *
 * ```ebnf
 * (* Introduced in 0.7.0 *)
 * UnicodeStringLiterals = (* item: *) UnicodeStringLiteral+;
 * ```
 */
export declare class UnicodeStringLiterals {
    /**
     * The underlying nonterminal CST node of kind `UnicodeStringLiterals`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `UnicodeStringLiterals`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `UnicodeStringLiterals`.
     */
    cst: NonterminalNode);
    /**
     * Returns an array of the child nodes that have the label `item`.
     */
    get items(): readonly UnicodeStringLiteral[];
}
/**
 * This node represents a `YulStatements` nonterminal, with the following structure:
 *
 * ```ebnf
 * YulStatements = (* item: *) YulStatement*;
 * ```
 */
export declare class YulStatements {
    /**
     * The underlying nonterminal CST node of kind `YulStatements`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `YulStatements`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `YulStatements`.
     */
    cst: NonterminalNode);
    /**
     * Returns an array of the child nodes that have the label `item`.
     */
    get items(): readonly YulStatement[];
}
/**
 * This node represents a `YulSwitchCases` nonterminal, with the following structure:
 *
 * ```ebnf
 * YulSwitchCases = (* item: *) YulSwitchCase+;
 * ```
 */
export declare class YulSwitchCases {
    /**
     * The underlying nonterminal CST node of kind `YulSwitchCases`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `YulSwitchCases`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `YulSwitchCases`.
     */
    cst: NonterminalNode);
    /**
     * Returns an array of the child nodes that have the label `item`.
     */
    get items(): readonly YulSwitchCase[];
}
/**
 * This node represents a `VersionExpressionSets` nonterminal, with the following structure:
 *
 * ```ebnf
 * VersionExpressionSets = (* item: *) VersionExpressionSet ((* separator: *) BAR_BAR (* item: *) VersionExpressionSet)*;
 * ```
 */
export declare class VersionExpressionSets {
    /**
     * The underlying nonterminal CST node of kind `VersionExpressionSets`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `VersionExpressionSets`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `VersionExpressionSets`.
     */
    cst: NonterminalNode);
    /**
     * Returns an array of the child nodes that have the label `item`.
     */
    get items(): readonly VersionExpressionSet[];
    /**
     * Returns an array of the child nodes that have the label `separator`.
     */
    get separators(): readonly TerminalNode[];
}
/**
 * This node represents a `SimpleVersionLiteral` nonterminal, with the following structure:
 *
 * ```ebnf
 * SimpleVersionLiteral = (* item: *) VERSION_SPECIFIER ((* separator: *) PERIOD (* item: *) VERSION_SPECIFIER)*;
 * ```
 */
export declare class SimpleVersionLiteral {
    /**
     * The underlying nonterminal CST node of kind `SimpleVersionLiteral`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `SimpleVersionLiteral`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `SimpleVersionLiteral`.
     */
    cst: NonterminalNode);
    /**
     * Returns an array of the child nodes that have the label `item`.
     */
    get items(): readonly TerminalNode[];
    /**
     * Returns an array of the child nodes that have the label `separator`.
     */
    get separators(): readonly TerminalNode[];
}
/**
 * This node represents a `ImportDeconstructionSymbols` nonterminal, with the following structure:
 *
 * ```ebnf
 * ImportDeconstructionSymbols = (* item: *) ImportDeconstructionSymbol ((* separator: *) COMMA (* item: *) ImportDeconstructionSymbol)*;
 * ```
 */
export declare class ImportDeconstructionSymbols {
    /**
     * The underlying nonterminal CST node of kind `ImportDeconstructionSymbols`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `ImportDeconstructionSymbols`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `ImportDeconstructionSymbols`.
     */
    cst: NonterminalNode);
    /**
     * Returns an array of the child nodes that have the label `item`.
     */
    get items(): readonly ImportDeconstructionSymbol[];
    /**
     * Returns an array of the child nodes that have the label `separator`.
     */
    get separators(): readonly TerminalNode[];
}
/**
 * This node represents a `UsingDeconstructionSymbols` nonterminal, with the following structure:
 *
 * ```ebnf
 * (* Introduced in 0.8.13 *)
 * UsingDeconstructionSymbols = (* item: *) UsingDeconstructionSymbol ((* separator: *) COMMA (* item: *) UsingDeconstructionSymbol)*;
 * ```
 */
export declare class UsingDeconstructionSymbols {
    /**
     * The underlying nonterminal CST node of kind `UsingDeconstructionSymbols`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `UsingDeconstructionSymbols`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `UsingDeconstructionSymbols`.
     */
    cst: NonterminalNode);
    /**
     * Returns an array of the child nodes that have the label `item`.
     */
    get items(): readonly UsingDeconstructionSymbol[];
    /**
     * Returns an array of the child nodes that have the label `separator`.
     */
    get separators(): readonly TerminalNode[];
}
/**
 * This node represents a `InheritanceTypes` nonterminal, with the following structure:
 *
 * ```ebnf
 * InheritanceTypes = (* item: *) InheritanceType ((* separator: *) COMMA (* item: *) InheritanceType)*;
 * ```
 */
export declare class InheritanceTypes {
    /**
     * The underlying nonterminal CST node of kind `InheritanceTypes`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `InheritanceTypes`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `InheritanceTypes`.
     */
    cst: NonterminalNode);
    /**
     * Returns an array of the child nodes that have the label `item`.
     */
    get items(): readonly InheritanceType[];
    /**
     * Returns an array of the child nodes that have the label `separator`.
     */
    get separators(): readonly TerminalNode[];
}
/**
 * This node represents a `EnumMembers` nonterminal, with the following structure:
 *
 * ```ebnf
 * EnumMembers = ((* item: *) IDENTIFIER ((* separator: *) COMMA (* item: *) IDENTIFIER)*)?;
 * ```
 */
export declare class EnumMembers {
    /**
     * The underlying nonterminal CST node of kind `EnumMembers`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `EnumMembers`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `EnumMembers`.
     */
    cst: NonterminalNode);
    /**
     * Returns an array of the child nodes that have the label `item`.
     */
    get items(): readonly TerminalNode[];
    /**
     * Returns an array of the child nodes that have the label `separator`.
     */
    get separators(): readonly TerminalNode[];
}
/**
 * This node represents a `Parameters` nonterminal, with the following structure:
 *
 * ```ebnf
 * Parameters = ((* item: *) Parameter ((* separator: *) COMMA (* item: *) Parameter)*)?;
 * ```
 */
export declare class Parameters {
    /**
     * The underlying nonterminal CST node of kind `Parameters`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `Parameters`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `Parameters`.
     */
    cst: NonterminalNode);
    /**
     * Returns an array of the child nodes that have the label `item`.
     */
    get items(): readonly Parameter[];
    /**
     * Returns an array of the child nodes that have the label `separator`.
     */
    get separators(): readonly TerminalNode[];
}
/**
 * This node represents a `OverridePaths` nonterminal, with the following structure:
 *
 * ```ebnf
 * (* Introduced in 0.6.0 *)
 * OverridePaths = (* item: *) IdentifierPath ((* separator: *) COMMA (* item: *) IdentifierPath)*;
 * ```
 */
export declare class OverridePaths {
    /**
     * The underlying nonterminal CST node of kind `OverridePaths`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `OverridePaths`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `OverridePaths`.
     */
    cst: NonterminalNode);
    /**
     * Returns an array of the child nodes that have the label `item`.
     */
    get items(): readonly IdentifierPath[];
    /**
     * Returns an array of the child nodes that have the label `separator`.
     */
    get separators(): readonly TerminalNode[];
}
/**
 * This node represents a `EventParameters` nonterminal, with the following structure:
 *
 * ```ebnf
 * EventParameters = ((* item: *) EventParameter ((* separator: *) COMMA (* item: *) EventParameter)*)?;
 * ```
 */
export declare class EventParameters {
    /**
     * The underlying nonterminal CST node of kind `EventParameters`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `EventParameters`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `EventParameters`.
     */
    cst: NonterminalNode);
    /**
     * Returns an array of the child nodes that have the label `item`.
     */
    get items(): readonly EventParameter[];
    /**
     * Returns an array of the child nodes that have the label `separator`.
     */
    get separators(): readonly TerminalNode[];
}
/**
 * This node represents a `ErrorParameters` nonterminal, with the following structure:
 *
 * ```ebnf
 * (* Introduced in 0.8.4 *)
 * ErrorParameters = ((* item: *) ErrorParameter ((* separator: *) COMMA (* item: *) ErrorParameter)*)?;
 * ```
 */
export declare class ErrorParameters {
    /**
     * The underlying nonterminal CST node of kind `ErrorParameters`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `ErrorParameters`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `ErrorParameters`.
     */
    cst: NonterminalNode);
    /**
     * Returns an array of the child nodes that have the label `item`.
     */
    get items(): readonly ErrorParameter[];
    /**
     * Returns an array of the child nodes that have the label `separator`.
     */
    get separators(): readonly TerminalNode[];
}
/**
 * This node represents a `AssemblyFlags` nonterminal, with the following structure:
 *
 * ```ebnf
 * (* Introduced in 0.8.13 *)
 * AssemblyFlags = (* item: *) StringLiteral ((* separator: *) COMMA (* item: *) StringLiteral)*;
 * ```
 */
export declare class AssemblyFlags {
    /**
     * The underlying nonterminal CST node of kind `AssemblyFlags`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `AssemblyFlags`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `AssemblyFlags`.
     */
    cst: NonterminalNode);
    /**
     * Returns an array of the child nodes that have the label `item`.
     */
    get items(): readonly StringLiteral[];
    /**
     * Returns an array of the child nodes that have the label `separator`.
     */
    get separators(): readonly TerminalNode[];
}
/**
 * This node represents a `TupleDeconstructionElements` nonterminal, with the following structure:
 *
 * ```ebnf
 * TupleDeconstructionElements = (* item: *) TupleDeconstructionElement ((* separator: *) COMMA (* item: *) TupleDeconstructionElement)*;
 * ```
 */
export declare class TupleDeconstructionElements {
    /**
     * The underlying nonterminal CST node of kind `TupleDeconstructionElements`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `TupleDeconstructionElements`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `TupleDeconstructionElements`.
     */
    cst: NonterminalNode);
    /**
     * Returns an array of the child nodes that have the label `item`.
     */
    get items(): readonly TupleDeconstructionElement[];
    /**
     * Returns an array of the child nodes that have the label `separator`.
     */
    get separators(): readonly TerminalNode[];
}
/**
 * This node represents a `PositionalArguments` nonterminal, with the following structure:
 *
 * ```ebnf
 * PositionalArguments = ((* item: *) Expression ((* separator: *) COMMA (* item: *) Expression)*)?;
 * ```
 */
export declare class PositionalArguments {
    /**
     * The underlying nonterminal CST node of kind `PositionalArguments`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `PositionalArguments`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `PositionalArguments`.
     */
    cst: NonterminalNode);
    /**
     * Returns an array of the child nodes that have the label `item`.
     */
    get items(): readonly Expression[];
    /**
     * Returns an array of the child nodes that have the label `separator`.
     */
    get separators(): readonly TerminalNode[];
}
/**
 * This node represents a `NamedArguments` nonterminal, with the following structure:
 *
 * ```ebnf
 * NamedArguments = ((* item: *) NamedArgument ((* separator: *) COMMA (* item: *) NamedArgument)*)?;
 * ```
 */
export declare class NamedArguments {
    /**
     * The underlying nonterminal CST node of kind `NamedArguments`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `NamedArguments`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `NamedArguments`.
     */
    cst: NonterminalNode);
    /**
     * Returns an array of the child nodes that have the label `item`.
     */
    get items(): readonly NamedArgument[];
    /**
     * Returns an array of the child nodes that have the label `separator`.
     */
    get separators(): readonly TerminalNode[];
}
/**
 * This node represents a `CallOptions` nonterminal, with the following structure:
 *
 * ```ebnf
 * (* Introduced in 0.6.2 *)
 * CallOptions = (* item: *) NamedArgument ((* separator: *) COMMA (* item: *) NamedArgument)*;
 * ```
 */
export declare class CallOptions {
    /**
     * The underlying nonterminal CST node of kind `CallOptions`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `CallOptions`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `CallOptions`.
     */
    cst: NonterminalNode);
    /**
     * Returns an array of the child nodes that have the label `item`.
     */
    get items(): readonly NamedArgument[];
    /**
     * Returns an array of the child nodes that have the label `separator`.
     */
    get separators(): readonly TerminalNode[];
}
/**
 * This node represents a `TupleValues` nonterminal, with the following structure:
 *
 * ```ebnf
 * TupleValues = (* item: *) TupleValue ((* separator: *) COMMA (* item: *) TupleValue)*;
 * ```
 */
export declare class TupleValues {
    /**
     * The underlying nonterminal CST node of kind `TupleValues`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `TupleValues`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `TupleValues`.
     */
    cst: NonterminalNode);
    /**
     * Returns an array of the child nodes that have the label `item`.
     */
    get items(): readonly TupleValue[];
    /**
     * Returns an array of the child nodes that have the label `separator`.
     */
    get separators(): readonly TerminalNode[];
}
/**
 * This node represents a `ArrayValues` nonterminal, with the following structure:
 *
 * ```ebnf
 * ArrayValues = (* item: *) Expression ((* separator: *) COMMA (* item: *) Expression)*;
 * ```
 */
export declare class ArrayValues {
    /**
     * The underlying nonterminal CST node of kind `ArrayValues`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `ArrayValues`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `ArrayValues`.
     */
    cst: NonterminalNode);
    /**
     * Returns an array of the child nodes that have the label `item`.
     */
    get items(): readonly Expression[];
    /**
     * Returns an array of the child nodes that have the label `separator`.
     */
    get separators(): readonly TerminalNode[];
}
/**
 * This node represents a `IdentifierPath` nonterminal, with the following structure:
 *
 * ```ebnf
 * IdentifierPath = (* item: *) IDENTIFIER ((* separator: *) PERIOD (* item: *) IDENTIFIER)*;
 * ```
 */
export declare class IdentifierPath {
    /**
     * The underlying nonterminal CST node of kind `IdentifierPath`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `IdentifierPath`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `IdentifierPath`.
     */
    cst: NonterminalNode);
    /**
     * Returns an array of the child nodes that have the label `item`.
     */
    get items(): readonly TerminalNode[];
    /**
     * Returns an array of the child nodes that have the label `separator`.
     */
    get separators(): readonly TerminalNode[];
}
/**
 * This node represents a `YulParameters` nonterminal, with the following structure:
 *
 * ```ebnf
 * YulParameters = ((* item: *) YUL_IDENTIFIER ((* separator: *) COMMA (* item: *) YUL_IDENTIFIER)*)?;
 * ```
 */
export declare class YulParameters {
    /**
     * The underlying nonterminal CST node of kind `YulParameters`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `YulParameters`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `YulParameters`.
     */
    cst: NonterminalNode);
    /**
     * Returns an array of the child nodes that have the label `item`.
     */
    get items(): readonly TerminalNode[];
    /**
     * Returns an array of the child nodes that have the label `separator`.
     */
    get separators(): readonly TerminalNode[];
}
/**
 * This node represents a `YulVariableNames` nonterminal, with the following structure:
 *
 * ```ebnf
 * YulVariableNames = (* item: *) YUL_IDENTIFIER ((* separator: *) COMMA (* item: *) YUL_IDENTIFIER)*;
 * ```
 */
export declare class YulVariableNames {
    /**
     * The underlying nonterminal CST node of kind `YulVariableNames`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `YulVariableNames`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `YulVariableNames`.
     */
    cst: NonterminalNode);
    /**
     * Returns an array of the child nodes that have the label `item`.
     */
    get items(): readonly TerminalNode[];
    /**
     * Returns an array of the child nodes that have the label `separator`.
     */
    get separators(): readonly TerminalNode[];
}
/**
 * This node represents a `YulArguments` nonterminal, with the following structure:
 *
 * ```ebnf
 * YulArguments = ((* item: *) YulExpression ((* separator: *) COMMA (* item: *) YulExpression)*)?;
 * ```
 */
export declare class YulArguments {
    /**
     * The underlying nonterminal CST node of kind `YulArguments`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `YulArguments`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `YulArguments`.
     */
    cst: NonterminalNode);
    /**
     * Returns an array of the child nodes that have the label `item`.
     */
    get items(): readonly YulExpression[];
    /**
     * Returns an array of the child nodes that have the label `separator`.
     */
    get separators(): readonly TerminalNode[];
}
/**
 * This node represents a `YulPaths` nonterminal, with the following structure:
 *
 * ```ebnf
 * YulPaths = (* item: *) YulPath ((* separator: *) COMMA (* item: *) YulPath)*;
 * ```
 */
export declare class YulPaths {
    /**
     * The underlying nonterminal CST node of kind `YulPaths`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `YulPaths`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `YulPaths`.
     */
    cst: NonterminalNode);
    /**
     * Returns an array of the child nodes that have the label `item`.
     */
    get items(): readonly YulPath[];
    /**
     * Returns an array of the child nodes that have the label `separator`.
     */
    get separators(): readonly TerminalNode[];
}
/**
 * This node represents a `YulPath` nonterminal, with the following structure:
 *
 * ```ebnf
 * YulPath = (* item: *) YUL_IDENTIFIER ((* separator: *) PERIOD (* item: *) YUL_IDENTIFIER)*;
 * ```
 */
export declare class YulPath {
    /**
     * The underlying nonterminal CST node of kind `YulPath`.
     */
    readonly cst: NonterminalNode;
    private readonly fetch;
    /**
     * Constructs a new AST node of type `YulPath`, given a nonterminal CST node of the same kind.
     */
    constructor(
    /**
     * The underlying nonterminal CST node of kind `YulPath`.
     */
    cst: NonterminalNode);
    /**
     * Returns an array of the child nodes that have the label `item`.
     */
    get items(): readonly TerminalNode[];
    /**
     * Returns an array of the child nodes that have the label `separator`.
     */
    get separators(): readonly TerminalNode[];
}
//# sourceMappingURL=nodes.generated.d.mts.map