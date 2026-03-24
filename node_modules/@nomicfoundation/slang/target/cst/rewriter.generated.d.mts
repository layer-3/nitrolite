import { NonterminalNode, Node, TerminalNode } from "./index.mjs";
/** Abstract class to rewrite a CST. */
export declare abstract class BaseRewriter {
    /** Replaces the `node` with a new node. If the result is `undefined`, the node is removed from the tree.
        This function typically the entry point of the rewrite operation. */
    rewriteNode(node: Node): Node | undefined;
    /** Rewrites a non-terminal node. Typically called from `rewriteNode`. */
    rewriteNonterminalNode(node: NonterminalNode): Node | undefined;
    /** Rewrites a terminal node. Typically called from `rewriteNode`. */
    rewriteTerminalNode(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `AbicoderPragma` node, recursively traversing the children (unless overriden). */
    rewriteAbicoderPragma(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `AbicoderVersion` node, recursively traversing the children (unless overriden). */
    rewriteAbicoderVersion(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `AdditiveExpression` node, recursively traversing the children (unless overriden). */
    rewriteAdditiveExpression(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `AddressType` node, recursively traversing the children (unless overriden). */
    rewriteAddressType(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `AndExpression` node, recursively traversing the children (unless overriden). */
    rewriteAndExpression(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `ArgumentsDeclaration` node, recursively traversing the children (unless overriden). */
    rewriteArgumentsDeclaration(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `ArrayExpression` node, recursively traversing the children (unless overriden). */
    rewriteArrayExpression(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `ArrayTypeName` node, recursively traversing the children (unless overriden). */
    rewriteArrayTypeName(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `ArrayValues` node, recursively traversing the children (unless overriden). */
    rewriteArrayValues(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `AssemblyFlags` node, recursively traversing the children (unless overriden). */
    rewriteAssemblyFlags(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `AssemblyFlagsDeclaration` node, recursively traversing the children (unless overriden). */
    rewriteAssemblyFlagsDeclaration(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `AssemblyStatement` node, recursively traversing the children (unless overriden). */
    rewriteAssemblyStatement(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `AssignmentExpression` node, recursively traversing the children (unless overriden). */
    rewriteAssignmentExpression(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `BitwiseAndExpression` node, recursively traversing the children (unless overriden). */
    rewriteBitwiseAndExpression(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `BitwiseOrExpression` node, recursively traversing the children (unless overriden). */
    rewriteBitwiseOrExpression(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `BitwiseXorExpression` node, recursively traversing the children (unless overriden). */
    rewriteBitwiseXorExpression(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `Block` node, recursively traversing the children (unless overriden). */
    rewriteBlock(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `BreakStatement` node, recursively traversing the children (unless overriden). */
    rewriteBreakStatement(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `CallOptions` node, recursively traversing the children (unless overriden). */
    rewriteCallOptions(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `CallOptionsExpression` node, recursively traversing the children (unless overriden). */
    rewriteCallOptionsExpression(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `CatchClause` node, recursively traversing the children (unless overriden). */
    rewriteCatchClause(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `CatchClauseError` node, recursively traversing the children (unless overriden). */
    rewriteCatchClauseError(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `CatchClauses` node, recursively traversing the children (unless overriden). */
    rewriteCatchClauses(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `ConditionalExpression` node, recursively traversing the children (unless overriden). */
    rewriteConditionalExpression(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `ConstantDefinition` node, recursively traversing the children (unless overriden). */
    rewriteConstantDefinition(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `ConstructorAttribute` node, recursively traversing the children (unless overriden). */
    rewriteConstructorAttribute(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `ConstructorAttributes` node, recursively traversing the children (unless overriden). */
    rewriteConstructorAttributes(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `ConstructorDefinition` node, recursively traversing the children (unless overriden). */
    rewriteConstructorDefinition(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `ContinueStatement` node, recursively traversing the children (unless overriden). */
    rewriteContinueStatement(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `ContractDefinition` node, recursively traversing the children (unless overriden). */
    rewriteContractDefinition(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `ContractMember` node, recursively traversing the children (unless overriden). */
    rewriteContractMember(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `ContractMembers` node, recursively traversing the children (unless overriden). */
    rewriteContractMembers(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `ContractSpecifier` node, recursively traversing the children (unless overriden). */
    rewriteContractSpecifier(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `ContractSpecifiers` node, recursively traversing the children (unless overriden). */
    rewriteContractSpecifiers(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `DecimalNumberExpression` node, recursively traversing the children (unless overriden). */
    rewriteDecimalNumberExpression(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `DoWhileStatement` node, recursively traversing the children (unless overriden). */
    rewriteDoWhileStatement(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `ElementaryType` node, recursively traversing the children (unless overriden). */
    rewriteElementaryType(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `ElseBranch` node, recursively traversing the children (unless overriden). */
    rewriteElseBranch(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `EmitStatement` node, recursively traversing the children (unless overriden). */
    rewriteEmitStatement(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `EnumDefinition` node, recursively traversing the children (unless overriden). */
    rewriteEnumDefinition(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `EnumMembers` node, recursively traversing the children (unless overriden). */
    rewriteEnumMembers(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `EqualityExpression` node, recursively traversing the children (unless overriden). */
    rewriteEqualityExpression(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `ErrorDefinition` node, recursively traversing the children (unless overriden). */
    rewriteErrorDefinition(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `ErrorParameter` node, recursively traversing the children (unless overriden). */
    rewriteErrorParameter(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `ErrorParameters` node, recursively traversing the children (unless overriden). */
    rewriteErrorParameters(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `ErrorParametersDeclaration` node, recursively traversing the children (unless overriden). */
    rewriteErrorParametersDeclaration(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `EventDefinition` node, recursively traversing the children (unless overriden). */
    rewriteEventDefinition(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `EventParameter` node, recursively traversing the children (unless overriden). */
    rewriteEventParameter(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `EventParameters` node, recursively traversing the children (unless overriden). */
    rewriteEventParameters(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `EventParametersDeclaration` node, recursively traversing the children (unless overriden). */
    rewriteEventParametersDeclaration(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `ExperimentalFeature` node, recursively traversing the children (unless overriden). */
    rewriteExperimentalFeature(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `ExperimentalPragma` node, recursively traversing the children (unless overriden). */
    rewriteExperimentalPragma(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `ExponentiationExpression` node, recursively traversing the children (unless overriden). */
    rewriteExponentiationExpression(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `Expression` node, recursively traversing the children (unless overriden). */
    rewriteExpression(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `ExpressionStatement` node, recursively traversing the children (unless overriden). */
    rewriteExpressionStatement(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `FallbackFunctionAttribute` node, recursively traversing the children (unless overriden). */
    rewriteFallbackFunctionAttribute(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `FallbackFunctionAttributes` node, recursively traversing the children (unless overriden). */
    rewriteFallbackFunctionAttributes(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `FallbackFunctionDefinition` node, recursively traversing the children (unless overriden). */
    rewriteFallbackFunctionDefinition(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `ForStatement` node, recursively traversing the children (unless overriden). */
    rewriteForStatement(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `ForStatementCondition` node, recursively traversing the children (unless overriden). */
    rewriteForStatementCondition(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `ForStatementInitialization` node, recursively traversing the children (unless overriden). */
    rewriteForStatementInitialization(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `FunctionAttribute` node, recursively traversing the children (unless overriden). */
    rewriteFunctionAttribute(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `FunctionAttributes` node, recursively traversing the children (unless overriden). */
    rewriteFunctionAttributes(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `FunctionBody` node, recursively traversing the children (unless overriden). */
    rewriteFunctionBody(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `FunctionCallExpression` node, recursively traversing the children (unless overriden). */
    rewriteFunctionCallExpression(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `FunctionDefinition` node, recursively traversing the children (unless overriden). */
    rewriteFunctionDefinition(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `FunctionName` node, recursively traversing the children (unless overriden). */
    rewriteFunctionName(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `FunctionType` node, recursively traversing the children (unless overriden). */
    rewriteFunctionType(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `FunctionTypeAttribute` node, recursively traversing the children (unless overriden). */
    rewriteFunctionTypeAttribute(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `FunctionTypeAttributes` node, recursively traversing the children (unless overriden). */
    rewriteFunctionTypeAttributes(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `HexNumberExpression` node, recursively traversing the children (unless overriden). */
    rewriteHexNumberExpression(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `HexStringLiteral` node, recursively traversing the children (unless overriden). */
    rewriteHexStringLiteral(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `HexStringLiterals` node, recursively traversing the children (unless overriden). */
    rewriteHexStringLiterals(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `IdentifierPath` node, recursively traversing the children (unless overriden). */
    rewriteIdentifierPath(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `IfStatement` node, recursively traversing the children (unless overriden). */
    rewriteIfStatement(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `ImportAlias` node, recursively traversing the children (unless overriden). */
    rewriteImportAlias(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `ImportClause` node, recursively traversing the children (unless overriden). */
    rewriteImportClause(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `ImportDeconstruction` node, recursively traversing the children (unless overriden). */
    rewriteImportDeconstruction(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `ImportDeconstructionSymbol` node, recursively traversing the children (unless overriden). */
    rewriteImportDeconstructionSymbol(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `ImportDeconstructionSymbols` node, recursively traversing the children (unless overriden). */
    rewriteImportDeconstructionSymbols(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `ImportDirective` node, recursively traversing the children (unless overriden). */
    rewriteImportDirective(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `IndexAccessEnd` node, recursively traversing the children (unless overriden). */
    rewriteIndexAccessEnd(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `IndexAccessExpression` node, recursively traversing the children (unless overriden). */
    rewriteIndexAccessExpression(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `InequalityExpression` node, recursively traversing the children (unless overriden). */
    rewriteInequalityExpression(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `InheritanceSpecifier` node, recursively traversing the children (unless overriden). */
    rewriteInheritanceSpecifier(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `InheritanceType` node, recursively traversing the children (unless overriden). */
    rewriteInheritanceType(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `InheritanceTypes` node, recursively traversing the children (unless overriden). */
    rewriteInheritanceTypes(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `InterfaceDefinition` node, recursively traversing the children (unless overriden). */
    rewriteInterfaceDefinition(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `InterfaceMembers` node, recursively traversing the children (unless overriden). */
    rewriteInterfaceMembers(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `LibraryDefinition` node, recursively traversing the children (unless overriden). */
    rewriteLibraryDefinition(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `LibraryMembers` node, recursively traversing the children (unless overriden). */
    rewriteLibraryMembers(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `MappingKey` node, recursively traversing the children (unless overriden). */
    rewriteMappingKey(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `MappingKeyType` node, recursively traversing the children (unless overriden). */
    rewriteMappingKeyType(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `MappingType` node, recursively traversing the children (unless overriden). */
    rewriteMappingType(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `MappingValue` node, recursively traversing the children (unless overriden). */
    rewriteMappingValue(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `MemberAccessExpression` node, recursively traversing the children (unless overriden). */
    rewriteMemberAccessExpression(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `ModifierAttribute` node, recursively traversing the children (unless overriden). */
    rewriteModifierAttribute(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `ModifierAttributes` node, recursively traversing the children (unless overriden). */
    rewriteModifierAttributes(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `ModifierDefinition` node, recursively traversing the children (unless overriden). */
    rewriteModifierDefinition(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `ModifierInvocation` node, recursively traversing the children (unless overriden). */
    rewriteModifierInvocation(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `MultiplicativeExpression` node, recursively traversing the children (unless overriden). */
    rewriteMultiplicativeExpression(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `NamedArgument` node, recursively traversing the children (unless overriden). */
    rewriteNamedArgument(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `NamedArgumentGroup` node, recursively traversing the children (unless overriden). */
    rewriteNamedArgumentGroup(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `NamedArguments` node, recursively traversing the children (unless overriden). */
    rewriteNamedArguments(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `NamedArgumentsDeclaration` node, recursively traversing the children (unless overriden). */
    rewriteNamedArgumentsDeclaration(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `NamedImport` node, recursively traversing the children (unless overriden). */
    rewriteNamedImport(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `NewExpression` node, recursively traversing the children (unless overriden). */
    rewriteNewExpression(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `NumberUnit` node, recursively traversing the children (unless overriden). */
    rewriteNumberUnit(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `OrExpression` node, recursively traversing the children (unless overriden). */
    rewriteOrExpression(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `OverridePaths` node, recursively traversing the children (unless overriden). */
    rewriteOverridePaths(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `OverridePathsDeclaration` node, recursively traversing the children (unless overriden). */
    rewriteOverridePathsDeclaration(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `OverrideSpecifier` node, recursively traversing the children (unless overriden). */
    rewriteOverrideSpecifier(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `Parameter` node, recursively traversing the children (unless overriden). */
    rewriteParameter(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `Parameters` node, recursively traversing the children (unless overriden). */
    rewriteParameters(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `ParametersDeclaration` node, recursively traversing the children (unless overriden). */
    rewriteParametersDeclaration(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `PathImport` node, recursively traversing the children (unless overriden). */
    rewritePathImport(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `PositionalArguments` node, recursively traversing the children (unless overriden). */
    rewritePositionalArguments(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `PositionalArgumentsDeclaration` node, recursively traversing the children (unless overriden). */
    rewritePositionalArgumentsDeclaration(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `PostfixExpression` node, recursively traversing the children (unless overriden). */
    rewritePostfixExpression(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `Pragma` node, recursively traversing the children (unless overriden). */
    rewritePragma(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `PragmaDirective` node, recursively traversing the children (unless overriden). */
    rewritePragmaDirective(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `PrefixExpression` node, recursively traversing the children (unless overriden). */
    rewritePrefixExpression(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `ReceiveFunctionAttribute` node, recursively traversing the children (unless overriden). */
    rewriteReceiveFunctionAttribute(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `ReceiveFunctionAttributes` node, recursively traversing the children (unless overriden). */
    rewriteReceiveFunctionAttributes(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `ReceiveFunctionDefinition` node, recursively traversing the children (unless overriden). */
    rewriteReceiveFunctionDefinition(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `ReturnStatement` node, recursively traversing the children (unless overriden). */
    rewriteReturnStatement(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `ReturnsDeclaration` node, recursively traversing the children (unless overriden). */
    rewriteReturnsDeclaration(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `RevertStatement` node, recursively traversing the children (unless overriden). */
    rewriteRevertStatement(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `ShiftExpression` node, recursively traversing the children (unless overriden). */
    rewriteShiftExpression(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `SimpleVersionLiteral` node, recursively traversing the children (unless overriden). */
    rewriteSimpleVersionLiteral(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `SourceUnit` node, recursively traversing the children (unless overriden). */
    rewriteSourceUnit(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `SourceUnitMember` node, recursively traversing the children (unless overriden). */
    rewriteSourceUnitMember(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `SourceUnitMembers` node, recursively traversing the children (unless overriden). */
    rewriteSourceUnitMembers(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `StateVariableAttribute` node, recursively traversing the children (unless overriden). */
    rewriteStateVariableAttribute(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `StateVariableAttributes` node, recursively traversing the children (unless overriden). */
    rewriteStateVariableAttributes(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `StateVariableDefinition` node, recursively traversing the children (unless overriden). */
    rewriteStateVariableDefinition(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `StateVariableDefinitionValue` node, recursively traversing the children (unless overriden). */
    rewriteStateVariableDefinitionValue(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `Statement` node, recursively traversing the children (unless overriden). */
    rewriteStatement(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `Statements` node, recursively traversing the children (unless overriden). */
    rewriteStatements(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `StorageLayoutSpecifier` node, recursively traversing the children (unless overriden). */
    rewriteStorageLayoutSpecifier(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `StorageLocation` node, recursively traversing the children (unless overriden). */
    rewriteStorageLocation(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `StringExpression` node, recursively traversing the children (unless overriden). */
    rewriteStringExpression(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `StringLiteral` node, recursively traversing the children (unless overriden). */
    rewriteStringLiteral(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `StringLiterals` node, recursively traversing the children (unless overriden). */
    rewriteStringLiterals(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `StructDefinition` node, recursively traversing the children (unless overriden). */
    rewriteStructDefinition(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `StructMember` node, recursively traversing the children (unless overriden). */
    rewriteStructMember(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `StructMembers` node, recursively traversing the children (unless overriden). */
    rewriteStructMembers(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `ThrowStatement` node, recursively traversing the children (unless overriden). */
    rewriteThrowStatement(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `TryStatement` node, recursively traversing the children (unless overriden). */
    rewriteTryStatement(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `TupleDeconstructionElement` node, recursively traversing the children (unless overriden). */
    rewriteTupleDeconstructionElement(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `TupleDeconstructionElements` node, recursively traversing the children (unless overriden). */
    rewriteTupleDeconstructionElements(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `TupleDeconstructionStatement` node, recursively traversing the children (unless overriden). */
    rewriteTupleDeconstructionStatement(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `TupleExpression` node, recursively traversing the children (unless overriden). */
    rewriteTupleExpression(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `TupleMember` node, recursively traversing the children (unless overriden). */
    rewriteTupleMember(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `TupleValue` node, recursively traversing the children (unless overriden). */
    rewriteTupleValue(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `TupleValues` node, recursively traversing the children (unless overriden). */
    rewriteTupleValues(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `TypeExpression` node, recursively traversing the children (unless overriden). */
    rewriteTypeExpression(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `TypeName` node, recursively traversing the children (unless overriden). */
    rewriteTypeName(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `TypedTupleMember` node, recursively traversing the children (unless overriden). */
    rewriteTypedTupleMember(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `UncheckedBlock` node, recursively traversing the children (unless overriden). */
    rewriteUncheckedBlock(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `UnicodeStringLiteral` node, recursively traversing the children (unless overriden). */
    rewriteUnicodeStringLiteral(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `UnicodeStringLiterals` node, recursively traversing the children (unless overriden). */
    rewriteUnicodeStringLiterals(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `UnnamedFunctionAttribute` node, recursively traversing the children (unless overriden). */
    rewriteUnnamedFunctionAttribute(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `UnnamedFunctionAttributes` node, recursively traversing the children (unless overriden). */
    rewriteUnnamedFunctionAttributes(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `UnnamedFunctionDefinition` node, recursively traversing the children (unless overriden). */
    rewriteUnnamedFunctionDefinition(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `UntypedTupleMember` node, recursively traversing the children (unless overriden). */
    rewriteUntypedTupleMember(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `UserDefinedValueTypeDefinition` node, recursively traversing the children (unless overriden). */
    rewriteUserDefinedValueTypeDefinition(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `UsingAlias` node, recursively traversing the children (unless overriden). */
    rewriteUsingAlias(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `UsingClause` node, recursively traversing the children (unless overriden). */
    rewriteUsingClause(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `UsingDeconstruction` node, recursively traversing the children (unless overriden). */
    rewriteUsingDeconstruction(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `UsingDeconstructionSymbol` node, recursively traversing the children (unless overriden). */
    rewriteUsingDeconstructionSymbol(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `UsingDeconstructionSymbols` node, recursively traversing the children (unless overriden). */
    rewriteUsingDeconstructionSymbols(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `UsingDirective` node, recursively traversing the children (unless overriden). */
    rewriteUsingDirective(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `UsingOperator` node, recursively traversing the children (unless overriden). */
    rewriteUsingOperator(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `UsingTarget` node, recursively traversing the children (unless overriden). */
    rewriteUsingTarget(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `VariableDeclarationStatement` node, recursively traversing the children (unless overriden). */
    rewriteVariableDeclarationStatement(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `VariableDeclarationType` node, recursively traversing the children (unless overriden). */
    rewriteVariableDeclarationType(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `VariableDeclarationValue` node, recursively traversing the children (unless overriden). */
    rewriteVariableDeclarationValue(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `VersionExpression` node, recursively traversing the children (unless overriden). */
    rewriteVersionExpression(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `VersionExpressionSet` node, recursively traversing the children (unless overriden). */
    rewriteVersionExpressionSet(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `VersionExpressionSets` node, recursively traversing the children (unless overriden). */
    rewriteVersionExpressionSets(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `VersionLiteral` node, recursively traversing the children (unless overriden). */
    rewriteVersionLiteral(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `VersionOperator` node, recursively traversing the children (unless overriden). */
    rewriteVersionOperator(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `VersionPragma` node, recursively traversing the children (unless overriden). */
    rewriteVersionPragma(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `VersionRange` node, recursively traversing the children (unless overriden). */
    rewriteVersionRange(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `VersionTerm` node, recursively traversing the children (unless overriden). */
    rewriteVersionTerm(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `WhileStatement` node, recursively traversing the children (unless overriden). */
    rewriteWhileStatement(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `YulArguments` node, recursively traversing the children (unless overriden). */
    rewriteYulArguments(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `YulAssignmentOperator` node, recursively traversing the children (unless overriden). */
    rewriteYulAssignmentOperator(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `YulBlock` node, recursively traversing the children (unless overriden). */
    rewriteYulBlock(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `YulBreakStatement` node, recursively traversing the children (unless overriden). */
    rewriteYulBreakStatement(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `YulColonAndEqual` node, recursively traversing the children (unless overriden). */
    rewriteYulColonAndEqual(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `YulContinueStatement` node, recursively traversing the children (unless overriden). */
    rewriteYulContinueStatement(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `YulDefaultCase` node, recursively traversing the children (unless overriden). */
    rewriteYulDefaultCase(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `YulEqualAndColon` node, recursively traversing the children (unless overriden). */
    rewriteYulEqualAndColon(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `YulExpression` node, recursively traversing the children (unless overriden). */
    rewriteYulExpression(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `YulForStatement` node, recursively traversing the children (unless overriden). */
    rewriteYulForStatement(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `YulFunctionCallExpression` node, recursively traversing the children (unless overriden). */
    rewriteYulFunctionCallExpression(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `YulFunctionDefinition` node, recursively traversing the children (unless overriden). */
    rewriteYulFunctionDefinition(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `YulIfStatement` node, recursively traversing the children (unless overriden). */
    rewriteYulIfStatement(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `YulLabel` node, recursively traversing the children (unless overriden). */
    rewriteYulLabel(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `YulLeaveStatement` node, recursively traversing the children (unless overriden). */
    rewriteYulLeaveStatement(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `YulLiteral` node, recursively traversing the children (unless overriden). */
    rewriteYulLiteral(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `YulParameters` node, recursively traversing the children (unless overriden). */
    rewriteYulParameters(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `YulParametersDeclaration` node, recursively traversing the children (unless overriden). */
    rewriteYulParametersDeclaration(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `YulPath` node, recursively traversing the children (unless overriden). */
    rewriteYulPath(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `YulPaths` node, recursively traversing the children (unless overriden). */
    rewriteYulPaths(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `YulReturnsDeclaration` node, recursively traversing the children (unless overriden). */
    rewriteYulReturnsDeclaration(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `YulStackAssignmentOperator` node, recursively traversing the children (unless overriden). */
    rewriteYulStackAssignmentOperator(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `YulStackAssignmentStatement` node, recursively traversing the children (unless overriden). */
    rewriteYulStackAssignmentStatement(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `YulStatement` node, recursively traversing the children (unless overriden). */
    rewriteYulStatement(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `YulStatements` node, recursively traversing the children (unless overriden). */
    rewriteYulStatements(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `YulSwitchCase` node, recursively traversing the children (unless overriden). */
    rewriteYulSwitchCase(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `YulSwitchCases` node, recursively traversing the children (unless overriden). */
    rewriteYulSwitchCases(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `YulSwitchStatement` node, recursively traversing the children (unless overriden). */
    rewriteYulSwitchStatement(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `YulValueCase` node, recursively traversing the children (unless overriden). */
    rewriteYulValueCase(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `YulVariableAssignmentStatement` node, recursively traversing the children (unless overriden). */
    rewriteYulVariableAssignmentStatement(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `YulVariableDeclarationStatement` node, recursively traversing the children (unless overriden). */
    rewriteYulVariableDeclarationStatement(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `YulVariableDeclarationValue` node, recursively traversing the children (unless overriden). */
    rewriteYulVariableDeclarationValue(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `YulVariableNames` node, recursively traversing the children (unless overriden). */
    rewriteYulVariableNames(node: NonterminalNode): Node | undefined;
    /** @virtual Rewrites a `ABIEncoderV2Keyword` node. */
    rewriteABIEncoderV2Keyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `AbicoderKeyword` node. */
    rewriteAbicoderKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `AbicoderV1Keyword` node. */
    rewriteAbicoderV1Keyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `AbicoderV2Keyword` node. */
    rewriteAbicoderV2Keyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `AbstractKeyword` node. */
    rewriteAbstractKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `AddressKeyword` node. */
    rewriteAddressKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `AfterKeyword` node. */
    rewriteAfterKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `AliasKeyword` node. */
    rewriteAliasKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `Ampersand` node. */
    rewriteAmpersand(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `AmpersandAmpersand` node. */
    rewriteAmpersandAmpersand(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `AmpersandEqual` node. */
    rewriteAmpersandEqual(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `AnonymousKeyword` node. */
    rewriteAnonymousKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `ApplyKeyword` node. */
    rewriteApplyKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `AsKeyword` node. */
    rewriteAsKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `AssemblyKeyword` node. */
    rewriteAssemblyKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `Asterisk` node. */
    rewriteAsterisk(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `AsteriskAsterisk` node. */
    rewriteAsteriskAsterisk(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `AsteriskEqual` node. */
    rewriteAsteriskEqual(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `AtKeyword` node. */
    rewriteAtKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `AutoKeyword` node. */
    rewriteAutoKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `Bang` node. */
    rewriteBang(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `BangEqual` node. */
    rewriteBangEqual(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `Bar` node. */
    rewriteBar(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `BarBar` node. */
    rewriteBarBar(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `BarEqual` node. */
    rewriteBarEqual(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `BoolKeyword` node. */
    rewriteBoolKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `BreakKeyword` node. */
    rewriteBreakKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `ByteKeyword` node. */
    rewriteByteKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `BytesKeyword` node. */
    rewriteBytesKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `CallDataKeyword` node. */
    rewriteCallDataKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `Caret` node. */
    rewriteCaret(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `CaretEqual` node. */
    rewriteCaretEqual(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `CaseKeyword` node. */
    rewriteCaseKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `CatchKeyword` node. */
    rewriteCatchKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `CloseBrace` node. */
    rewriteCloseBrace(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `CloseBracket` node. */
    rewriteCloseBracket(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `CloseParen` node. */
    rewriteCloseParen(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `Colon` node. */
    rewriteColon(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `ColonEqual` node. */
    rewriteColonEqual(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `Comma` node. */
    rewriteComma(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `ConstantKeyword` node. */
    rewriteConstantKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `ConstructorKeyword` node. */
    rewriteConstructorKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `ContinueKeyword` node. */
    rewriteContinueKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `ContractKeyword` node. */
    rewriteContractKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `CopyOfKeyword` node. */
    rewriteCopyOfKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `DaysKeyword` node. */
    rewriteDaysKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `DecimalLiteral` node. */
    rewriteDecimalLiteral(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `DefaultKeyword` node. */
    rewriteDefaultKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `DefineKeyword` node. */
    rewriteDefineKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `DeleteKeyword` node. */
    rewriteDeleteKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `DoKeyword` node. */
    rewriteDoKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `DoubleQuotedHexStringLiteral` node. */
    rewriteDoubleQuotedHexStringLiteral(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `DoubleQuotedStringLiteral` node. */
    rewriteDoubleQuotedStringLiteral(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `DoubleQuotedUnicodeStringLiteral` node. */
    rewriteDoubleQuotedUnicodeStringLiteral(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `DoubleQuotedVersionLiteral` node. */
    rewriteDoubleQuotedVersionLiteral(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `ElseKeyword` node. */
    rewriteElseKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `EmitKeyword` node. */
    rewriteEmitKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `EndOfLine` node. */
    rewriteEndOfLine(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `EnumKeyword` node. */
    rewriteEnumKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `Equal` node. */
    rewriteEqual(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `EqualColon` node. */
    rewriteEqualColon(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `EqualEqual` node. */
    rewriteEqualEqual(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `EqualGreaterThan` node. */
    rewriteEqualGreaterThan(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `ErrorKeyword` node. */
    rewriteErrorKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `EtherKeyword` node. */
    rewriteEtherKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `EventKeyword` node. */
    rewriteEventKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `ExperimentalKeyword` node. */
    rewriteExperimentalKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `ExternalKeyword` node. */
    rewriteExternalKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `FallbackKeyword` node. */
    rewriteFallbackKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `FalseKeyword` node. */
    rewriteFalseKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `FinalKeyword` node. */
    rewriteFinalKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `FinneyKeyword` node. */
    rewriteFinneyKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `FixedKeyword` node. */
    rewriteFixedKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `ForKeyword` node. */
    rewriteForKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `FromKeyword` node. */
    rewriteFromKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `FunctionKeyword` node. */
    rewriteFunctionKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `GlobalKeyword` node. */
    rewriteGlobalKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `GreaterThan` node. */
    rewriteGreaterThan(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `GreaterThanEqual` node. */
    rewriteGreaterThanEqual(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `GreaterThanGreaterThan` node. */
    rewriteGreaterThanGreaterThan(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `GreaterThanGreaterThanEqual` node. */
    rewriteGreaterThanGreaterThanEqual(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `GreaterThanGreaterThanGreaterThan` node. */
    rewriteGreaterThanGreaterThanGreaterThan(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `GreaterThanGreaterThanGreaterThanEqual` node. */
    rewriteGreaterThanGreaterThanGreaterThanEqual(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `GweiKeyword` node. */
    rewriteGweiKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `HexKeyword` node. */
    rewriteHexKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `HexLiteral` node. */
    rewriteHexLiteral(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `HoursKeyword` node. */
    rewriteHoursKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `Identifier` node. */
    rewriteIdentifier(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `IfKeyword` node. */
    rewriteIfKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `ImmutableKeyword` node. */
    rewriteImmutableKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `ImplementsKeyword` node. */
    rewriteImplementsKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `ImportKeyword` node. */
    rewriteImportKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `InKeyword` node. */
    rewriteInKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `IndexedKeyword` node. */
    rewriteIndexedKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `InlineKeyword` node. */
    rewriteInlineKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `IntKeyword` node. */
    rewriteIntKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `InterfaceKeyword` node. */
    rewriteInterfaceKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `InternalKeyword` node. */
    rewriteInternalKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `IsKeyword` node. */
    rewriteIsKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `LayoutKeyword` node. */
    rewriteLayoutKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `LessThan` node. */
    rewriteLessThan(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `LessThanEqual` node. */
    rewriteLessThanEqual(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `LessThanLessThan` node. */
    rewriteLessThanLessThan(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `LessThanLessThanEqual` node. */
    rewriteLessThanLessThanEqual(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `LetKeyword` node. */
    rewriteLetKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `LibraryKeyword` node. */
    rewriteLibraryKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `MacroKeyword` node. */
    rewriteMacroKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `MappingKeyword` node. */
    rewriteMappingKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `MatchKeyword` node. */
    rewriteMatchKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `MemoryKeyword` node. */
    rewriteMemoryKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `Minus` node. */
    rewriteMinus(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `MinusEqual` node. */
    rewriteMinusEqual(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `MinusGreaterThan` node. */
    rewriteMinusGreaterThan(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `MinusMinus` node. */
    rewriteMinusMinus(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `MinutesKeyword` node. */
    rewriteMinutesKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `ModifierKeyword` node. */
    rewriteModifierKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `MultiLineComment` node. */
    rewriteMultiLineComment(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `MultiLineNatSpecComment` node. */
    rewriteMultiLineNatSpecComment(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `MutableKeyword` node. */
    rewriteMutableKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `NewKeyword` node. */
    rewriteNewKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `NullKeyword` node. */
    rewriteNullKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `OfKeyword` node. */
    rewriteOfKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `OpenBrace` node. */
    rewriteOpenBrace(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `OpenBracket` node. */
    rewriteOpenBracket(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `OpenParen` node. */
    rewriteOpenParen(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `OverrideKeyword` node. */
    rewriteOverrideKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `PartialKeyword` node. */
    rewritePartialKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `PayableKeyword` node. */
    rewritePayableKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `Percent` node. */
    rewritePercent(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `PercentEqual` node. */
    rewritePercentEqual(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `Period` node. */
    rewritePeriod(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `Plus` node. */
    rewritePlus(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `PlusEqual` node. */
    rewritePlusEqual(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `PlusPlus` node. */
    rewritePlusPlus(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `PragmaKeyword` node. */
    rewritePragmaKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `PrivateKeyword` node. */
    rewritePrivateKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `PromiseKeyword` node. */
    rewritePromiseKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `PublicKeyword` node. */
    rewritePublicKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `PureKeyword` node. */
    rewritePureKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `QuestionMark` node. */
    rewriteQuestionMark(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `ReceiveKeyword` node. */
    rewriteReceiveKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `ReferenceKeyword` node. */
    rewriteReferenceKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `RelocatableKeyword` node. */
    rewriteRelocatableKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `ReturnKeyword` node. */
    rewriteReturnKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `ReturnsKeyword` node. */
    rewriteReturnsKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `RevertKeyword` node. */
    rewriteRevertKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `SMTCheckerKeyword` node. */
    rewriteSMTCheckerKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `SealedKeyword` node. */
    rewriteSealedKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `SecondsKeyword` node. */
    rewriteSecondsKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `Semicolon` node. */
    rewriteSemicolon(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `SingleLineComment` node. */
    rewriteSingleLineComment(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `SingleLineNatSpecComment` node. */
    rewriteSingleLineNatSpecComment(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `SingleQuotedHexStringLiteral` node. */
    rewriteSingleQuotedHexStringLiteral(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `SingleQuotedStringLiteral` node. */
    rewriteSingleQuotedStringLiteral(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `SingleQuotedUnicodeStringLiteral` node. */
    rewriteSingleQuotedUnicodeStringLiteral(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `SingleQuotedVersionLiteral` node. */
    rewriteSingleQuotedVersionLiteral(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `SizeOfKeyword` node. */
    rewriteSizeOfKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `Slash` node. */
    rewriteSlash(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `SlashEqual` node. */
    rewriteSlashEqual(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `SolidityKeyword` node. */
    rewriteSolidityKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `StaticKeyword` node. */
    rewriteStaticKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `StorageKeyword` node. */
    rewriteStorageKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `StringKeyword` node. */
    rewriteStringKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `StructKeyword` node. */
    rewriteStructKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `SuperKeyword` node. */
    rewriteSuperKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `SupportsKeyword` node. */
    rewriteSupportsKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `SwitchKeyword` node. */
    rewriteSwitchKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `SzaboKeyword` node. */
    rewriteSzaboKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `ThisKeyword` node. */
    rewriteThisKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `ThrowKeyword` node. */
    rewriteThrowKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `Tilde` node. */
    rewriteTilde(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `TransientKeyword` node. */
    rewriteTransientKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `TrueKeyword` node. */
    rewriteTrueKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `TryKeyword` node. */
    rewriteTryKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `TypeDefKeyword` node. */
    rewriteTypeDefKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `TypeKeyword` node. */
    rewriteTypeKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `TypeOfKeyword` node. */
    rewriteTypeOfKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `UfixedKeyword` node. */
    rewriteUfixedKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `UintKeyword` node. */
    rewriteUintKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `UncheckedKeyword` node. */
    rewriteUncheckedKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `UsingKeyword` node. */
    rewriteUsingKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `VarKeyword` node. */
    rewriteVarKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `VersionSpecifier` node. */
    rewriteVersionSpecifier(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `ViewKeyword` node. */
    rewriteViewKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `VirtualKeyword` node. */
    rewriteVirtualKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `WeeksKeyword` node. */
    rewriteWeeksKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `WeiKeyword` node. */
    rewriteWeiKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `WhileKeyword` node. */
    rewriteWhileKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `Whitespace` node. */
    rewriteWhitespace(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `YearsKeyword` node. */
    rewriteYearsKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `YulAbstractKeyword` node. */
    rewriteYulAbstractKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `YulAfterKeyword` node. */
    rewriteYulAfterKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `YulAliasKeyword` node. */
    rewriteYulAliasKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `YulAnonymousKeyword` node. */
    rewriteYulAnonymousKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `YulApplyKeyword` node. */
    rewriteYulApplyKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `YulAsKeyword` node. */
    rewriteYulAsKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `YulAssemblyKeyword` node. */
    rewriteYulAssemblyKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `YulAutoKeyword` node. */
    rewriteYulAutoKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `YulBoolKeyword` node. */
    rewriteYulBoolKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `YulBreakKeyword` node. */
    rewriteYulBreakKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `YulBytesKeyword` node. */
    rewriteYulBytesKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `YulCallDataKeyword` node. */
    rewriteYulCallDataKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `YulCaseKeyword` node. */
    rewriteYulCaseKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `YulCatchKeyword` node. */
    rewriteYulCatchKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `YulConstantKeyword` node. */
    rewriteYulConstantKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `YulConstructorKeyword` node. */
    rewriteYulConstructorKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `YulContinueKeyword` node. */
    rewriteYulContinueKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `YulContractKeyword` node. */
    rewriteYulContractKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `YulCopyOfKeyword` node. */
    rewriteYulCopyOfKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `YulDaysKeyword` node. */
    rewriteYulDaysKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `YulDecimalLiteral` node. */
    rewriteYulDecimalLiteral(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `YulDefaultKeyword` node. */
    rewriteYulDefaultKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `YulDefineKeyword` node. */
    rewriteYulDefineKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `YulDeleteKeyword` node. */
    rewriteYulDeleteKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `YulDoKeyword` node. */
    rewriteYulDoKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `YulElseKeyword` node. */
    rewriteYulElseKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `YulEmitKeyword` node. */
    rewriteYulEmitKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `YulEnumKeyword` node. */
    rewriteYulEnumKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `YulEtherKeyword` node. */
    rewriteYulEtherKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `YulEventKeyword` node. */
    rewriteYulEventKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `YulExternalKeyword` node. */
    rewriteYulExternalKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `YulFallbackKeyword` node. */
    rewriteYulFallbackKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `YulFalseKeyword` node. */
    rewriteYulFalseKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `YulFinalKeyword` node. */
    rewriteYulFinalKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `YulFinneyKeyword` node. */
    rewriteYulFinneyKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `YulFixedKeyword` node. */
    rewriteYulFixedKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `YulForKeyword` node. */
    rewriteYulForKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `YulFunctionKeyword` node. */
    rewriteYulFunctionKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `YulGweiKeyword` node. */
    rewriteYulGweiKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `YulHexKeyword` node. */
    rewriteYulHexKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `YulHexLiteral` node. */
    rewriteYulHexLiteral(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `YulHoursKeyword` node. */
    rewriteYulHoursKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `YulIdentifier` node. */
    rewriteYulIdentifier(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `YulIfKeyword` node. */
    rewriteYulIfKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `YulImmutableKeyword` node. */
    rewriteYulImmutableKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `YulImplementsKeyword` node. */
    rewriteYulImplementsKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `YulImportKeyword` node. */
    rewriteYulImportKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `YulInKeyword` node. */
    rewriteYulInKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `YulIndexedKeyword` node. */
    rewriteYulIndexedKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `YulInlineKeyword` node. */
    rewriteYulInlineKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `YulIntKeyword` node. */
    rewriteYulIntKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `YulInterfaceKeyword` node. */
    rewriteYulInterfaceKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `YulInternalKeyword` node. */
    rewriteYulInternalKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `YulIsKeyword` node. */
    rewriteYulIsKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `YulLeaveKeyword` node. */
    rewriteYulLeaveKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `YulLetKeyword` node. */
    rewriteYulLetKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `YulLibraryKeyword` node. */
    rewriteYulLibraryKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `YulMacroKeyword` node. */
    rewriteYulMacroKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `YulMappingKeyword` node. */
    rewriteYulMappingKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `YulMatchKeyword` node. */
    rewriteYulMatchKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `YulMemoryKeyword` node. */
    rewriteYulMemoryKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `YulMinutesKeyword` node. */
    rewriteYulMinutesKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `YulModifierKeyword` node. */
    rewriteYulModifierKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `YulMutableKeyword` node. */
    rewriteYulMutableKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `YulNewKeyword` node. */
    rewriteYulNewKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `YulNullKeyword` node. */
    rewriteYulNullKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `YulOfKeyword` node. */
    rewriteYulOfKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `YulOverrideKeyword` node. */
    rewriteYulOverrideKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `YulPartialKeyword` node. */
    rewriteYulPartialKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `YulPayableKeyword` node. */
    rewriteYulPayableKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `YulPragmaKeyword` node. */
    rewriteYulPragmaKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `YulPrivateKeyword` node. */
    rewriteYulPrivateKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `YulPromiseKeyword` node. */
    rewriteYulPromiseKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `YulPublicKeyword` node. */
    rewriteYulPublicKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `YulPureKeyword` node. */
    rewriteYulPureKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `YulReceiveKeyword` node. */
    rewriteYulReceiveKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `YulReferenceKeyword` node. */
    rewriteYulReferenceKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `YulRelocatableKeyword` node. */
    rewriteYulRelocatableKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `YulReturnsKeyword` node. */
    rewriteYulReturnsKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `YulSealedKeyword` node. */
    rewriteYulSealedKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `YulSecondsKeyword` node. */
    rewriteYulSecondsKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `YulSizeOfKeyword` node. */
    rewriteYulSizeOfKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `YulStaticKeyword` node. */
    rewriteYulStaticKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `YulStorageKeyword` node. */
    rewriteYulStorageKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `YulStringKeyword` node. */
    rewriteYulStringKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `YulStructKeyword` node. */
    rewriteYulStructKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `YulSuperKeyword` node. */
    rewriteYulSuperKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `YulSupportsKeyword` node. */
    rewriteYulSupportsKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `YulSwitchKeyword` node. */
    rewriteYulSwitchKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `YulSzaboKeyword` node. */
    rewriteYulSzaboKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `YulThisKeyword` node. */
    rewriteYulThisKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `YulThrowKeyword` node. */
    rewriteYulThrowKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `YulTrueKeyword` node. */
    rewriteYulTrueKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `YulTryKeyword` node. */
    rewriteYulTryKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `YulTypeDefKeyword` node. */
    rewriteYulTypeDefKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `YulTypeKeyword` node. */
    rewriteYulTypeKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `YulTypeOfKeyword` node. */
    rewriteYulTypeOfKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `YulUfixedKeyword` node. */
    rewriteYulUfixedKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `YulUintKeyword` node. */
    rewriteYulUintKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `YulUncheckedKeyword` node. */
    rewriteYulUncheckedKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `YulUsingKeyword` node. */
    rewriteYulUsingKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `YulVarKeyword` node. */
    rewriteYulVarKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `YulViewKeyword` node. */
    rewriteYulViewKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `YulVirtualKeyword` node. */
    rewriteYulVirtualKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `YulWeeksKeyword` node. */
    rewriteYulWeeksKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `YulWeiKeyword` node. */
    rewriteYulWeiKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `YulWhileKeyword` node. */
    rewriteYulWhileKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `YulYearsKeyword` node. */
    rewriteYulYearsKeyword(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites an `Unrecognized` node. */
    rewriteUnrecognized(node: TerminalNode): Node | undefined;
    /** @virtual Rewrites a `Missing` node. */
    rewriteMissing(node: TerminalNode): Node | undefined;
    /** Rewrites all the children of a given non-terminal node. */
    protected rewriteChildren(node: NonterminalNode): NonterminalNode;
}
//# sourceMappingURL=rewriter.generated.d.mts.map