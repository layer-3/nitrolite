// This file is generated automatically by infrastructure scripts. Please don't edit by hand.
import { NonterminalKind, NonterminalNode, NodeType, TerminalKind, Edge } from "./index.mjs";
/** Abstract class to rewrite a CST. */
export class BaseRewriter {
    /** Replaces the `node` with a new node. If the result is `undefined`, the node is removed from the tree.
        This function typically the entry point of the rewrite operation. */
    rewriteNode(node) {
        switch (node.type) {
            case NodeType.TerminalNode:
                return this.rewriteTerminalNode(node);
            case NodeType.NonterminalNode:
                return this.rewriteNonterminalNode(node);
        }
    }
    /** Rewrites a non-terminal node. Typically called from `rewriteNode`. */
    rewriteNonterminalNode(node) {
        switch (node.kind) {
            case NonterminalKind.AbicoderPragma:
                return this.rewriteAbicoderPragma(node);
            case NonterminalKind.AbicoderVersion:
                return this.rewriteAbicoderVersion(node);
            case NonterminalKind.AdditiveExpression:
                return this.rewriteAdditiveExpression(node);
            case NonterminalKind.AddressType:
                return this.rewriteAddressType(node);
            case NonterminalKind.AndExpression:
                return this.rewriteAndExpression(node);
            case NonterminalKind.ArgumentsDeclaration:
                return this.rewriteArgumentsDeclaration(node);
            case NonterminalKind.ArrayExpression:
                return this.rewriteArrayExpression(node);
            case NonterminalKind.ArrayTypeName:
                return this.rewriteArrayTypeName(node);
            case NonterminalKind.ArrayValues:
                return this.rewriteArrayValues(node);
            case NonterminalKind.AssemblyFlags:
                return this.rewriteAssemblyFlags(node);
            case NonterminalKind.AssemblyFlagsDeclaration:
                return this.rewriteAssemblyFlagsDeclaration(node);
            case NonterminalKind.AssemblyStatement:
                return this.rewriteAssemblyStatement(node);
            case NonterminalKind.AssignmentExpression:
                return this.rewriteAssignmentExpression(node);
            case NonterminalKind.BitwiseAndExpression:
                return this.rewriteBitwiseAndExpression(node);
            case NonterminalKind.BitwiseOrExpression:
                return this.rewriteBitwiseOrExpression(node);
            case NonterminalKind.BitwiseXorExpression:
                return this.rewriteBitwiseXorExpression(node);
            case NonterminalKind.Block:
                return this.rewriteBlock(node);
            case NonterminalKind.BreakStatement:
                return this.rewriteBreakStatement(node);
            case NonterminalKind.CallOptions:
                return this.rewriteCallOptions(node);
            case NonterminalKind.CallOptionsExpression:
                return this.rewriteCallOptionsExpression(node);
            case NonterminalKind.CatchClause:
                return this.rewriteCatchClause(node);
            case NonterminalKind.CatchClauseError:
                return this.rewriteCatchClauseError(node);
            case NonterminalKind.CatchClauses:
                return this.rewriteCatchClauses(node);
            case NonterminalKind.ConditionalExpression:
                return this.rewriteConditionalExpression(node);
            case NonterminalKind.ConstantDefinition:
                return this.rewriteConstantDefinition(node);
            case NonterminalKind.ConstructorAttribute:
                return this.rewriteConstructorAttribute(node);
            case NonterminalKind.ConstructorAttributes:
                return this.rewriteConstructorAttributes(node);
            case NonterminalKind.ConstructorDefinition:
                return this.rewriteConstructorDefinition(node);
            case NonterminalKind.ContinueStatement:
                return this.rewriteContinueStatement(node);
            case NonterminalKind.ContractDefinition:
                return this.rewriteContractDefinition(node);
            case NonterminalKind.ContractMember:
                return this.rewriteContractMember(node);
            case NonterminalKind.ContractMembers:
                return this.rewriteContractMembers(node);
            case NonterminalKind.ContractSpecifier:
                return this.rewriteContractSpecifier(node);
            case NonterminalKind.ContractSpecifiers:
                return this.rewriteContractSpecifiers(node);
            case NonterminalKind.DecimalNumberExpression:
                return this.rewriteDecimalNumberExpression(node);
            case NonterminalKind.DoWhileStatement:
                return this.rewriteDoWhileStatement(node);
            case NonterminalKind.ElementaryType:
                return this.rewriteElementaryType(node);
            case NonterminalKind.ElseBranch:
                return this.rewriteElseBranch(node);
            case NonterminalKind.EmitStatement:
                return this.rewriteEmitStatement(node);
            case NonterminalKind.EnumDefinition:
                return this.rewriteEnumDefinition(node);
            case NonterminalKind.EnumMembers:
                return this.rewriteEnumMembers(node);
            case NonterminalKind.EqualityExpression:
                return this.rewriteEqualityExpression(node);
            case NonterminalKind.ErrorDefinition:
                return this.rewriteErrorDefinition(node);
            case NonterminalKind.ErrorParameter:
                return this.rewriteErrorParameter(node);
            case NonterminalKind.ErrorParameters:
                return this.rewriteErrorParameters(node);
            case NonterminalKind.ErrorParametersDeclaration:
                return this.rewriteErrorParametersDeclaration(node);
            case NonterminalKind.EventDefinition:
                return this.rewriteEventDefinition(node);
            case NonterminalKind.EventParameter:
                return this.rewriteEventParameter(node);
            case NonterminalKind.EventParameters:
                return this.rewriteEventParameters(node);
            case NonterminalKind.EventParametersDeclaration:
                return this.rewriteEventParametersDeclaration(node);
            case NonterminalKind.ExperimentalFeature:
                return this.rewriteExperimentalFeature(node);
            case NonterminalKind.ExperimentalPragma:
                return this.rewriteExperimentalPragma(node);
            case NonterminalKind.ExponentiationExpression:
                return this.rewriteExponentiationExpression(node);
            case NonterminalKind.Expression:
                return this.rewriteExpression(node);
            case NonterminalKind.ExpressionStatement:
                return this.rewriteExpressionStatement(node);
            case NonterminalKind.FallbackFunctionAttribute:
                return this.rewriteFallbackFunctionAttribute(node);
            case NonterminalKind.FallbackFunctionAttributes:
                return this.rewriteFallbackFunctionAttributes(node);
            case NonterminalKind.FallbackFunctionDefinition:
                return this.rewriteFallbackFunctionDefinition(node);
            case NonterminalKind.ForStatement:
                return this.rewriteForStatement(node);
            case NonterminalKind.ForStatementCondition:
                return this.rewriteForStatementCondition(node);
            case NonterminalKind.ForStatementInitialization:
                return this.rewriteForStatementInitialization(node);
            case NonterminalKind.FunctionAttribute:
                return this.rewriteFunctionAttribute(node);
            case NonterminalKind.FunctionAttributes:
                return this.rewriteFunctionAttributes(node);
            case NonterminalKind.FunctionBody:
                return this.rewriteFunctionBody(node);
            case NonterminalKind.FunctionCallExpression:
                return this.rewriteFunctionCallExpression(node);
            case NonterminalKind.FunctionDefinition:
                return this.rewriteFunctionDefinition(node);
            case NonterminalKind.FunctionName:
                return this.rewriteFunctionName(node);
            case NonterminalKind.FunctionType:
                return this.rewriteFunctionType(node);
            case NonterminalKind.FunctionTypeAttribute:
                return this.rewriteFunctionTypeAttribute(node);
            case NonterminalKind.FunctionTypeAttributes:
                return this.rewriteFunctionTypeAttributes(node);
            case NonterminalKind.HexNumberExpression:
                return this.rewriteHexNumberExpression(node);
            case NonterminalKind.HexStringLiteral:
                return this.rewriteHexStringLiteral(node);
            case NonterminalKind.HexStringLiterals:
                return this.rewriteHexStringLiterals(node);
            case NonterminalKind.IdentifierPath:
                return this.rewriteIdentifierPath(node);
            case NonterminalKind.IfStatement:
                return this.rewriteIfStatement(node);
            case NonterminalKind.ImportAlias:
                return this.rewriteImportAlias(node);
            case NonterminalKind.ImportClause:
                return this.rewriteImportClause(node);
            case NonterminalKind.ImportDeconstruction:
                return this.rewriteImportDeconstruction(node);
            case NonterminalKind.ImportDeconstructionSymbol:
                return this.rewriteImportDeconstructionSymbol(node);
            case NonterminalKind.ImportDeconstructionSymbols:
                return this.rewriteImportDeconstructionSymbols(node);
            case NonterminalKind.ImportDirective:
                return this.rewriteImportDirective(node);
            case NonterminalKind.IndexAccessEnd:
                return this.rewriteIndexAccessEnd(node);
            case NonterminalKind.IndexAccessExpression:
                return this.rewriteIndexAccessExpression(node);
            case NonterminalKind.InequalityExpression:
                return this.rewriteInequalityExpression(node);
            case NonterminalKind.InheritanceSpecifier:
                return this.rewriteInheritanceSpecifier(node);
            case NonterminalKind.InheritanceType:
                return this.rewriteInheritanceType(node);
            case NonterminalKind.InheritanceTypes:
                return this.rewriteInheritanceTypes(node);
            case NonterminalKind.InterfaceDefinition:
                return this.rewriteInterfaceDefinition(node);
            case NonterminalKind.InterfaceMembers:
                return this.rewriteInterfaceMembers(node);
            case NonterminalKind.LibraryDefinition:
                return this.rewriteLibraryDefinition(node);
            case NonterminalKind.LibraryMembers:
                return this.rewriteLibraryMembers(node);
            case NonterminalKind.MappingKey:
                return this.rewriteMappingKey(node);
            case NonterminalKind.MappingKeyType:
                return this.rewriteMappingKeyType(node);
            case NonterminalKind.MappingType:
                return this.rewriteMappingType(node);
            case NonterminalKind.MappingValue:
                return this.rewriteMappingValue(node);
            case NonterminalKind.MemberAccessExpression:
                return this.rewriteMemberAccessExpression(node);
            case NonterminalKind.ModifierAttribute:
                return this.rewriteModifierAttribute(node);
            case NonterminalKind.ModifierAttributes:
                return this.rewriteModifierAttributes(node);
            case NonterminalKind.ModifierDefinition:
                return this.rewriteModifierDefinition(node);
            case NonterminalKind.ModifierInvocation:
                return this.rewriteModifierInvocation(node);
            case NonterminalKind.MultiplicativeExpression:
                return this.rewriteMultiplicativeExpression(node);
            case NonterminalKind.NamedArgument:
                return this.rewriteNamedArgument(node);
            case NonterminalKind.NamedArgumentGroup:
                return this.rewriteNamedArgumentGroup(node);
            case NonterminalKind.NamedArguments:
                return this.rewriteNamedArguments(node);
            case NonterminalKind.NamedArgumentsDeclaration:
                return this.rewriteNamedArgumentsDeclaration(node);
            case NonterminalKind.NamedImport:
                return this.rewriteNamedImport(node);
            case NonterminalKind.NewExpression:
                return this.rewriteNewExpression(node);
            case NonterminalKind.NumberUnit:
                return this.rewriteNumberUnit(node);
            case NonterminalKind.OrExpression:
                return this.rewriteOrExpression(node);
            case NonterminalKind.OverridePaths:
                return this.rewriteOverridePaths(node);
            case NonterminalKind.OverridePathsDeclaration:
                return this.rewriteOverridePathsDeclaration(node);
            case NonterminalKind.OverrideSpecifier:
                return this.rewriteOverrideSpecifier(node);
            case NonterminalKind.Parameter:
                return this.rewriteParameter(node);
            case NonterminalKind.Parameters:
                return this.rewriteParameters(node);
            case NonterminalKind.ParametersDeclaration:
                return this.rewriteParametersDeclaration(node);
            case NonterminalKind.PathImport:
                return this.rewritePathImport(node);
            case NonterminalKind.PositionalArguments:
                return this.rewritePositionalArguments(node);
            case NonterminalKind.PositionalArgumentsDeclaration:
                return this.rewritePositionalArgumentsDeclaration(node);
            case NonterminalKind.PostfixExpression:
                return this.rewritePostfixExpression(node);
            case NonterminalKind.Pragma:
                return this.rewritePragma(node);
            case NonterminalKind.PragmaDirective:
                return this.rewritePragmaDirective(node);
            case NonterminalKind.PrefixExpression:
                return this.rewritePrefixExpression(node);
            case NonterminalKind.ReceiveFunctionAttribute:
                return this.rewriteReceiveFunctionAttribute(node);
            case NonterminalKind.ReceiveFunctionAttributes:
                return this.rewriteReceiveFunctionAttributes(node);
            case NonterminalKind.ReceiveFunctionDefinition:
                return this.rewriteReceiveFunctionDefinition(node);
            case NonterminalKind.ReturnStatement:
                return this.rewriteReturnStatement(node);
            case NonterminalKind.ReturnsDeclaration:
                return this.rewriteReturnsDeclaration(node);
            case NonterminalKind.RevertStatement:
                return this.rewriteRevertStatement(node);
            case NonterminalKind.ShiftExpression:
                return this.rewriteShiftExpression(node);
            case NonterminalKind.SimpleVersionLiteral:
                return this.rewriteSimpleVersionLiteral(node);
            case NonterminalKind.SourceUnit:
                return this.rewriteSourceUnit(node);
            case NonterminalKind.SourceUnitMember:
                return this.rewriteSourceUnitMember(node);
            case NonterminalKind.SourceUnitMembers:
                return this.rewriteSourceUnitMembers(node);
            case NonterminalKind.StateVariableAttribute:
                return this.rewriteStateVariableAttribute(node);
            case NonterminalKind.StateVariableAttributes:
                return this.rewriteStateVariableAttributes(node);
            case NonterminalKind.StateVariableDefinition:
                return this.rewriteStateVariableDefinition(node);
            case NonterminalKind.StateVariableDefinitionValue:
                return this.rewriteStateVariableDefinitionValue(node);
            case NonterminalKind.Statement:
                return this.rewriteStatement(node);
            case NonterminalKind.Statements:
                return this.rewriteStatements(node);
            case NonterminalKind.StorageLayoutSpecifier:
                return this.rewriteStorageLayoutSpecifier(node);
            case NonterminalKind.StorageLocation:
                return this.rewriteStorageLocation(node);
            case NonterminalKind.StringExpression:
                return this.rewriteStringExpression(node);
            case NonterminalKind.StringLiteral:
                return this.rewriteStringLiteral(node);
            case NonterminalKind.StringLiterals:
                return this.rewriteStringLiterals(node);
            case NonterminalKind.StructDefinition:
                return this.rewriteStructDefinition(node);
            case NonterminalKind.StructMember:
                return this.rewriteStructMember(node);
            case NonterminalKind.StructMembers:
                return this.rewriteStructMembers(node);
            case NonterminalKind.ThrowStatement:
                return this.rewriteThrowStatement(node);
            case NonterminalKind.TryStatement:
                return this.rewriteTryStatement(node);
            case NonterminalKind.TupleDeconstructionElement:
                return this.rewriteTupleDeconstructionElement(node);
            case NonterminalKind.TupleDeconstructionElements:
                return this.rewriteTupleDeconstructionElements(node);
            case NonterminalKind.TupleDeconstructionStatement:
                return this.rewriteTupleDeconstructionStatement(node);
            case NonterminalKind.TupleExpression:
                return this.rewriteTupleExpression(node);
            case NonterminalKind.TupleMember:
                return this.rewriteTupleMember(node);
            case NonterminalKind.TupleValue:
                return this.rewriteTupleValue(node);
            case NonterminalKind.TupleValues:
                return this.rewriteTupleValues(node);
            case NonterminalKind.TypeExpression:
                return this.rewriteTypeExpression(node);
            case NonterminalKind.TypeName:
                return this.rewriteTypeName(node);
            case NonterminalKind.TypedTupleMember:
                return this.rewriteTypedTupleMember(node);
            case NonterminalKind.UncheckedBlock:
                return this.rewriteUncheckedBlock(node);
            case NonterminalKind.UnicodeStringLiteral:
                return this.rewriteUnicodeStringLiteral(node);
            case NonterminalKind.UnicodeStringLiterals:
                return this.rewriteUnicodeStringLiterals(node);
            case NonterminalKind.UnnamedFunctionAttribute:
                return this.rewriteUnnamedFunctionAttribute(node);
            case NonterminalKind.UnnamedFunctionAttributes:
                return this.rewriteUnnamedFunctionAttributes(node);
            case NonterminalKind.UnnamedFunctionDefinition:
                return this.rewriteUnnamedFunctionDefinition(node);
            case NonterminalKind.UntypedTupleMember:
                return this.rewriteUntypedTupleMember(node);
            case NonterminalKind.UserDefinedValueTypeDefinition:
                return this.rewriteUserDefinedValueTypeDefinition(node);
            case NonterminalKind.UsingAlias:
                return this.rewriteUsingAlias(node);
            case NonterminalKind.UsingClause:
                return this.rewriteUsingClause(node);
            case NonterminalKind.UsingDeconstruction:
                return this.rewriteUsingDeconstruction(node);
            case NonterminalKind.UsingDeconstructionSymbol:
                return this.rewriteUsingDeconstructionSymbol(node);
            case NonterminalKind.UsingDeconstructionSymbols:
                return this.rewriteUsingDeconstructionSymbols(node);
            case NonterminalKind.UsingDirective:
                return this.rewriteUsingDirective(node);
            case NonterminalKind.UsingOperator:
                return this.rewriteUsingOperator(node);
            case NonterminalKind.UsingTarget:
                return this.rewriteUsingTarget(node);
            case NonterminalKind.VariableDeclarationStatement:
                return this.rewriteVariableDeclarationStatement(node);
            case NonterminalKind.VariableDeclarationType:
                return this.rewriteVariableDeclarationType(node);
            case NonterminalKind.VariableDeclarationValue:
                return this.rewriteVariableDeclarationValue(node);
            case NonterminalKind.VersionExpression:
                return this.rewriteVersionExpression(node);
            case NonterminalKind.VersionExpressionSet:
                return this.rewriteVersionExpressionSet(node);
            case NonterminalKind.VersionExpressionSets:
                return this.rewriteVersionExpressionSets(node);
            case NonterminalKind.VersionLiteral:
                return this.rewriteVersionLiteral(node);
            case NonterminalKind.VersionOperator:
                return this.rewriteVersionOperator(node);
            case NonterminalKind.VersionPragma:
                return this.rewriteVersionPragma(node);
            case NonterminalKind.VersionRange:
                return this.rewriteVersionRange(node);
            case NonterminalKind.VersionTerm:
                return this.rewriteVersionTerm(node);
            case NonterminalKind.WhileStatement:
                return this.rewriteWhileStatement(node);
            case NonterminalKind.YulArguments:
                return this.rewriteYulArguments(node);
            case NonterminalKind.YulAssignmentOperator:
                return this.rewriteYulAssignmentOperator(node);
            case NonterminalKind.YulBlock:
                return this.rewriteYulBlock(node);
            case NonterminalKind.YulBreakStatement:
                return this.rewriteYulBreakStatement(node);
            case NonterminalKind.YulColonAndEqual:
                return this.rewriteYulColonAndEqual(node);
            case NonterminalKind.YulContinueStatement:
                return this.rewriteYulContinueStatement(node);
            case NonterminalKind.YulDefaultCase:
                return this.rewriteYulDefaultCase(node);
            case NonterminalKind.YulEqualAndColon:
                return this.rewriteYulEqualAndColon(node);
            case NonterminalKind.YulExpression:
                return this.rewriteYulExpression(node);
            case NonterminalKind.YulForStatement:
                return this.rewriteYulForStatement(node);
            case NonterminalKind.YulFunctionCallExpression:
                return this.rewriteYulFunctionCallExpression(node);
            case NonterminalKind.YulFunctionDefinition:
                return this.rewriteYulFunctionDefinition(node);
            case NonterminalKind.YulIfStatement:
                return this.rewriteYulIfStatement(node);
            case NonterminalKind.YulLabel:
                return this.rewriteYulLabel(node);
            case NonterminalKind.YulLeaveStatement:
                return this.rewriteYulLeaveStatement(node);
            case NonterminalKind.YulLiteral:
                return this.rewriteYulLiteral(node);
            case NonterminalKind.YulParameters:
                return this.rewriteYulParameters(node);
            case NonterminalKind.YulParametersDeclaration:
                return this.rewriteYulParametersDeclaration(node);
            case NonterminalKind.YulPath:
                return this.rewriteYulPath(node);
            case NonterminalKind.YulPaths:
                return this.rewriteYulPaths(node);
            case NonterminalKind.YulReturnsDeclaration:
                return this.rewriteYulReturnsDeclaration(node);
            case NonterminalKind.YulStackAssignmentOperator:
                return this.rewriteYulStackAssignmentOperator(node);
            case NonterminalKind.YulStackAssignmentStatement:
                return this.rewriteYulStackAssignmentStatement(node);
            case NonterminalKind.YulStatement:
                return this.rewriteYulStatement(node);
            case NonterminalKind.YulStatements:
                return this.rewriteYulStatements(node);
            case NonterminalKind.YulSwitchCase:
                return this.rewriteYulSwitchCase(node);
            case NonterminalKind.YulSwitchCases:
                return this.rewriteYulSwitchCases(node);
            case NonterminalKind.YulSwitchStatement:
                return this.rewriteYulSwitchStatement(node);
            case NonterminalKind.YulValueCase:
                return this.rewriteYulValueCase(node);
            case NonterminalKind.YulVariableAssignmentStatement:
                return this.rewriteYulVariableAssignmentStatement(node);
            case NonterminalKind.YulVariableDeclarationStatement:
                return this.rewriteYulVariableDeclarationStatement(node);
            case NonterminalKind.YulVariableDeclarationValue:
                return this.rewriteYulVariableDeclarationValue(node);
            case NonterminalKind.YulVariableNames:
                return this.rewriteYulVariableNames(node);
        }
    }
    /** Rewrites a terminal node. Typically called from `rewriteNode`. */
    rewriteTerminalNode(node) {
        switch (node.kind) {
            case TerminalKind.ABIEncoderV2Keyword:
                return this.rewriteABIEncoderV2Keyword(node);
            case TerminalKind.AbicoderKeyword:
                return this.rewriteAbicoderKeyword(node);
            case TerminalKind.AbicoderV1Keyword:
                return this.rewriteAbicoderV1Keyword(node);
            case TerminalKind.AbicoderV2Keyword:
                return this.rewriteAbicoderV2Keyword(node);
            case TerminalKind.AbstractKeyword:
                return this.rewriteAbstractKeyword(node);
            case TerminalKind.AddressKeyword:
                return this.rewriteAddressKeyword(node);
            case TerminalKind.AfterKeyword:
                return this.rewriteAfterKeyword(node);
            case TerminalKind.AliasKeyword:
                return this.rewriteAliasKeyword(node);
            case TerminalKind.Ampersand:
                return this.rewriteAmpersand(node);
            case TerminalKind.AmpersandAmpersand:
                return this.rewriteAmpersandAmpersand(node);
            case TerminalKind.AmpersandEqual:
                return this.rewriteAmpersandEqual(node);
            case TerminalKind.AnonymousKeyword:
                return this.rewriteAnonymousKeyword(node);
            case TerminalKind.ApplyKeyword:
                return this.rewriteApplyKeyword(node);
            case TerminalKind.AsKeyword:
                return this.rewriteAsKeyword(node);
            case TerminalKind.AssemblyKeyword:
                return this.rewriteAssemblyKeyword(node);
            case TerminalKind.Asterisk:
                return this.rewriteAsterisk(node);
            case TerminalKind.AsteriskAsterisk:
                return this.rewriteAsteriskAsterisk(node);
            case TerminalKind.AsteriskEqual:
                return this.rewriteAsteriskEqual(node);
            case TerminalKind.AtKeyword:
                return this.rewriteAtKeyword(node);
            case TerminalKind.AutoKeyword:
                return this.rewriteAutoKeyword(node);
            case TerminalKind.Bang:
                return this.rewriteBang(node);
            case TerminalKind.BangEqual:
                return this.rewriteBangEqual(node);
            case TerminalKind.Bar:
                return this.rewriteBar(node);
            case TerminalKind.BarBar:
                return this.rewriteBarBar(node);
            case TerminalKind.BarEqual:
                return this.rewriteBarEqual(node);
            case TerminalKind.BoolKeyword:
                return this.rewriteBoolKeyword(node);
            case TerminalKind.BreakKeyword:
                return this.rewriteBreakKeyword(node);
            case TerminalKind.ByteKeyword:
                return this.rewriteByteKeyword(node);
            case TerminalKind.BytesKeyword:
                return this.rewriteBytesKeyword(node);
            case TerminalKind.CallDataKeyword:
                return this.rewriteCallDataKeyword(node);
            case TerminalKind.Caret:
                return this.rewriteCaret(node);
            case TerminalKind.CaretEqual:
                return this.rewriteCaretEqual(node);
            case TerminalKind.CaseKeyword:
                return this.rewriteCaseKeyword(node);
            case TerminalKind.CatchKeyword:
                return this.rewriteCatchKeyword(node);
            case TerminalKind.CloseBrace:
                return this.rewriteCloseBrace(node);
            case TerminalKind.CloseBracket:
                return this.rewriteCloseBracket(node);
            case TerminalKind.CloseParen:
                return this.rewriteCloseParen(node);
            case TerminalKind.Colon:
                return this.rewriteColon(node);
            case TerminalKind.ColonEqual:
                return this.rewriteColonEqual(node);
            case TerminalKind.Comma:
                return this.rewriteComma(node);
            case TerminalKind.ConstantKeyword:
                return this.rewriteConstantKeyword(node);
            case TerminalKind.ConstructorKeyword:
                return this.rewriteConstructorKeyword(node);
            case TerminalKind.ContinueKeyword:
                return this.rewriteContinueKeyword(node);
            case TerminalKind.ContractKeyword:
                return this.rewriteContractKeyword(node);
            case TerminalKind.CopyOfKeyword:
                return this.rewriteCopyOfKeyword(node);
            case TerminalKind.DaysKeyword:
                return this.rewriteDaysKeyword(node);
            case TerminalKind.DecimalLiteral:
                return this.rewriteDecimalLiteral(node);
            case TerminalKind.DefaultKeyword:
                return this.rewriteDefaultKeyword(node);
            case TerminalKind.DefineKeyword:
                return this.rewriteDefineKeyword(node);
            case TerminalKind.DeleteKeyword:
                return this.rewriteDeleteKeyword(node);
            case TerminalKind.DoKeyword:
                return this.rewriteDoKeyword(node);
            case TerminalKind.DoubleQuotedHexStringLiteral:
                return this.rewriteDoubleQuotedHexStringLiteral(node);
            case TerminalKind.DoubleQuotedStringLiteral:
                return this.rewriteDoubleQuotedStringLiteral(node);
            case TerminalKind.DoubleQuotedUnicodeStringLiteral:
                return this.rewriteDoubleQuotedUnicodeStringLiteral(node);
            case TerminalKind.DoubleQuotedVersionLiteral:
                return this.rewriteDoubleQuotedVersionLiteral(node);
            case TerminalKind.ElseKeyword:
                return this.rewriteElseKeyword(node);
            case TerminalKind.EmitKeyword:
                return this.rewriteEmitKeyword(node);
            case TerminalKind.EndOfLine:
                return this.rewriteEndOfLine(node);
            case TerminalKind.EnumKeyword:
                return this.rewriteEnumKeyword(node);
            case TerminalKind.Equal:
                return this.rewriteEqual(node);
            case TerminalKind.EqualColon:
                return this.rewriteEqualColon(node);
            case TerminalKind.EqualEqual:
                return this.rewriteEqualEqual(node);
            case TerminalKind.EqualGreaterThan:
                return this.rewriteEqualGreaterThan(node);
            case TerminalKind.ErrorKeyword:
                return this.rewriteErrorKeyword(node);
            case TerminalKind.EtherKeyword:
                return this.rewriteEtherKeyword(node);
            case TerminalKind.EventKeyword:
                return this.rewriteEventKeyword(node);
            case TerminalKind.ExperimentalKeyword:
                return this.rewriteExperimentalKeyword(node);
            case TerminalKind.ExternalKeyword:
                return this.rewriteExternalKeyword(node);
            case TerminalKind.FallbackKeyword:
                return this.rewriteFallbackKeyword(node);
            case TerminalKind.FalseKeyword:
                return this.rewriteFalseKeyword(node);
            case TerminalKind.FinalKeyword:
                return this.rewriteFinalKeyword(node);
            case TerminalKind.FinneyKeyword:
                return this.rewriteFinneyKeyword(node);
            case TerminalKind.FixedKeyword:
                return this.rewriteFixedKeyword(node);
            case TerminalKind.ForKeyword:
                return this.rewriteForKeyword(node);
            case TerminalKind.FromKeyword:
                return this.rewriteFromKeyword(node);
            case TerminalKind.FunctionKeyword:
                return this.rewriteFunctionKeyword(node);
            case TerminalKind.GlobalKeyword:
                return this.rewriteGlobalKeyword(node);
            case TerminalKind.GreaterThan:
                return this.rewriteGreaterThan(node);
            case TerminalKind.GreaterThanEqual:
                return this.rewriteGreaterThanEqual(node);
            case TerminalKind.GreaterThanGreaterThan:
                return this.rewriteGreaterThanGreaterThan(node);
            case TerminalKind.GreaterThanGreaterThanEqual:
                return this.rewriteGreaterThanGreaterThanEqual(node);
            case TerminalKind.GreaterThanGreaterThanGreaterThan:
                return this.rewriteGreaterThanGreaterThanGreaterThan(node);
            case TerminalKind.GreaterThanGreaterThanGreaterThanEqual:
                return this.rewriteGreaterThanGreaterThanGreaterThanEqual(node);
            case TerminalKind.GweiKeyword:
                return this.rewriteGweiKeyword(node);
            case TerminalKind.HexKeyword:
                return this.rewriteHexKeyword(node);
            case TerminalKind.HexLiteral:
                return this.rewriteHexLiteral(node);
            case TerminalKind.HoursKeyword:
                return this.rewriteHoursKeyword(node);
            case TerminalKind.Identifier:
                return this.rewriteIdentifier(node);
            case TerminalKind.IfKeyword:
                return this.rewriteIfKeyword(node);
            case TerminalKind.ImmutableKeyword:
                return this.rewriteImmutableKeyword(node);
            case TerminalKind.ImplementsKeyword:
                return this.rewriteImplementsKeyword(node);
            case TerminalKind.ImportKeyword:
                return this.rewriteImportKeyword(node);
            case TerminalKind.InKeyword:
                return this.rewriteInKeyword(node);
            case TerminalKind.IndexedKeyword:
                return this.rewriteIndexedKeyword(node);
            case TerminalKind.InlineKeyword:
                return this.rewriteInlineKeyword(node);
            case TerminalKind.IntKeyword:
                return this.rewriteIntKeyword(node);
            case TerminalKind.InterfaceKeyword:
                return this.rewriteInterfaceKeyword(node);
            case TerminalKind.InternalKeyword:
                return this.rewriteInternalKeyword(node);
            case TerminalKind.IsKeyword:
                return this.rewriteIsKeyword(node);
            case TerminalKind.LayoutKeyword:
                return this.rewriteLayoutKeyword(node);
            case TerminalKind.LessThan:
                return this.rewriteLessThan(node);
            case TerminalKind.LessThanEqual:
                return this.rewriteLessThanEqual(node);
            case TerminalKind.LessThanLessThan:
                return this.rewriteLessThanLessThan(node);
            case TerminalKind.LessThanLessThanEqual:
                return this.rewriteLessThanLessThanEqual(node);
            case TerminalKind.LetKeyword:
                return this.rewriteLetKeyword(node);
            case TerminalKind.LibraryKeyword:
                return this.rewriteLibraryKeyword(node);
            case TerminalKind.MacroKeyword:
                return this.rewriteMacroKeyword(node);
            case TerminalKind.MappingKeyword:
                return this.rewriteMappingKeyword(node);
            case TerminalKind.MatchKeyword:
                return this.rewriteMatchKeyword(node);
            case TerminalKind.MemoryKeyword:
                return this.rewriteMemoryKeyword(node);
            case TerminalKind.Minus:
                return this.rewriteMinus(node);
            case TerminalKind.MinusEqual:
                return this.rewriteMinusEqual(node);
            case TerminalKind.MinusGreaterThan:
                return this.rewriteMinusGreaterThan(node);
            case TerminalKind.MinusMinus:
                return this.rewriteMinusMinus(node);
            case TerminalKind.MinutesKeyword:
                return this.rewriteMinutesKeyword(node);
            case TerminalKind.ModifierKeyword:
                return this.rewriteModifierKeyword(node);
            case TerminalKind.MultiLineComment:
                return this.rewriteMultiLineComment(node);
            case TerminalKind.MultiLineNatSpecComment:
                return this.rewriteMultiLineNatSpecComment(node);
            case TerminalKind.MutableKeyword:
                return this.rewriteMutableKeyword(node);
            case TerminalKind.NewKeyword:
                return this.rewriteNewKeyword(node);
            case TerminalKind.NullKeyword:
                return this.rewriteNullKeyword(node);
            case TerminalKind.OfKeyword:
                return this.rewriteOfKeyword(node);
            case TerminalKind.OpenBrace:
                return this.rewriteOpenBrace(node);
            case TerminalKind.OpenBracket:
                return this.rewriteOpenBracket(node);
            case TerminalKind.OpenParen:
                return this.rewriteOpenParen(node);
            case TerminalKind.OverrideKeyword:
                return this.rewriteOverrideKeyword(node);
            case TerminalKind.PartialKeyword:
                return this.rewritePartialKeyword(node);
            case TerminalKind.PayableKeyword:
                return this.rewritePayableKeyword(node);
            case TerminalKind.Percent:
                return this.rewritePercent(node);
            case TerminalKind.PercentEqual:
                return this.rewritePercentEqual(node);
            case TerminalKind.Period:
                return this.rewritePeriod(node);
            case TerminalKind.Plus:
                return this.rewritePlus(node);
            case TerminalKind.PlusEqual:
                return this.rewritePlusEqual(node);
            case TerminalKind.PlusPlus:
                return this.rewritePlusPlus(node);
            case TerminalKind.PragmaKeyword:
                return this.rewritePragmaKeyword(node);
            case TerminalKind.PrivateKeyword:
                return this.rewritePrivateKeyword(node);
            case TerminalKind.PromiseKeyword:
                return this.rewritePromiseKeyword(node);
            case TerminalKind.PublicKeyword:
                return this.rewritePublicKeyword(node);
            case TerminalKind.PureKeyword:
                return this.rewritePureKeyword(node);
            case TerminalKind.QuestionMark:
                return this.rewriteQuestionMark(node);
            case TerminalKind.ReceiveKeyword:
                return this.rewriteReceiveKeyword(node);
            case TerminalKind.ReferenceKeyword:
                return this.rewriteReferenceKeyword(node);
            case TerminalKind.RelocatableKeyword:
                return this.rewriteRelocatableKeyword(node);
            case TerminalKind.ReturnKeyword:
                return this.rewriteReturnKeyword(node);
            case TerminalKind.ReturnsKeyword:
                return this.rewriteReturnsKeyword(node);
            case TerminalKind.RevertKeyword:
                return this.rewriteRevertKeyword(node);
            case TerminalKind.SMTCheckerKeyword:
                return this.rewriteSMTCheckerKeyword(node);
            case TerminalKind.SealedKeyword:
                return this.rewriteSealedKeyword(node);
            case TerminalKind.SecondsKeyword:
                return this.rewriteSecondsKeyword(node);
            case TerminalKind.Semicolon:
                return this.rewriteSemicolon(node);
            case TerminalKind.SingleLineComment:
                return this.rewriteSingleLineComment(node);
            case TerminalKind.SingleLineNatSpecComment:
                return this.rewriteSingleLineNatSpecComment(node);
            case TerminalKind.SingleQuotedHexStringLiteral:
                return this.rewriteSingleQuotedHexStringLiteral(node);
            case TerminalKind.SingleQuotedStringLiteral:
                return this.rewriteSingleQuotedStringLiteral(node);
            case TerminalKind.SingleQuotedUnicodeStringLiteral:
                return this.rewriteSingleQuotedUnicodeStringLiteral(node);
            case TerminalKind.SingleQuotedVersionLiteral:
                return this.rewriteSingleQuotedVersionLiteral(node);
            case TerminalKind.SizeOfKeyword:
                return this.rewriteSizeOfKeyword(node);
            case TerminalKind.Slash:
                return this.rewriteSlash(node);
            case TerminalKind.SlashEqual:
                return this.rewriteSlashEqual(node);
            case TerminalKind.SolidityKeyword:
                return this.rewriteSolidityKeyword(node);
            case TerminalKind.StaticKeyword:
                return this.rewriteStaticKeyword(node);
            case TerminalKind.StorageKeyword:
                return this.rewriteStorageKeyword(node);
            case TerminalKind.StringKeyword:
                return this.rewriteStringKeyword(node);
            case TerminalKind.StructKeyword:
                return this.rewriteStructKeyword(node);
            case TerminalKind.SuperKeyword:
                return this.rewriteSuperKeyword(node);
            case TerminalKind.SupportsKeyword:
                return this.rewriteSupportsKeyword(node);
            case TerminalKind.SwitchKeyword:
                return this.rewriteSwitchKeyword(node);
            case TerminalKind.SzaboKeyword:
                return this.rewriteSzaboKeyword(node);
            case TerminalKind.ThisKeyword:
                return this.rewriteThisKeyword(node);
            case TerminalKind.ThrowKeyword:
                return this.rewriteThrowKeyword(node);
            case TerminalKind.Tilde:
                return this.rewriteTilde(node);
            case TerminalKind.TransientKeyword:
                return this.rewriteTransientKeyword(node);
            case TerminalKind.TrueKeyword:
                return this.rewriteTrueKeyword(node);
            case TerminalKind.TryKeyword:
                return this.rewriteTryKeyword(node);
            case TerminalKind.TypeDefKeyword:
                return this.rewriteTypeDefKeyword(node);
            case TerminalKind.TypeKeyword:
                return this.rewriteTypeKeyword(node);
            case TerminalKind.TypeOfKeyword:
                return this.rewriteTypeOfKeyword(node);
            case TerminalKind.UfixedKeyword:
                return this.rewriteUfixedKeyword(node);
            case TerminalKind.UintKeyword:
                return this.rewriteUintKeyword(node);
            case TerminalKind.UncheckedKeyword:
                return this.rewriteUncheckedKeyword(node);
            case TerminalKind.UsingKeyword:
                return this.rewriteUsingKeyword(node);
            case TerminalKind.VarKeyword:
                return this.rewriteVarKeyword(node);
            case TerminalKind.VersionSpecifier:
                return this.rewriteVersionSpecifier(node);
            case TerminalKind.ViewKeyword:
                return this.rewriteViewKeyword(node);
            case TerminalKind.VirtualKeyword:
                return this.rewriteVirtualKeyword(node);
            case TerminalKind.WeeksKeyword:
                return this.rewriteWeeksKeyword(node);
            case TerminalKind.WeiKeyword:
                return this.rewriteWeiKeyword(node);
            case TerminalKind.WhileKeyword:
                return this.rewriteWhileKeyword(node);
            case TerminalKind.Whitespace:
                return this.rewriteWhitespace(node);
            case TerminalKind.YearsKeyword:
                return this.rewriteYearsKeyword(node);
            case TerminalKind.YulAbstractKeyword:
                return this.rewriteYulAbstractKeyword(node);
            case TerminalKind.YulAfterKeyword:
                return this.rewriteYulAfterKeyword(node);
            case TerminalKind.YulAliasKeyword:
                return this.rewriteYulAliasKeyword(node);
            case TerminalKind.YulAnonymousKeyword:
                return this.rewriteYulAnonymousKeyword(node);
            case TerminalKind.YulApplyKeyword:
                return this.rewriteYulApplyKeyword(node);
            case TerminalKind.YulAsKeyword:
                return this.rewriteYulAsKeyword(node);
            case TerminalKind.YulAssemblyKeyword:
                return this.rewriteYulAssemblyKeyword(node);
            case TerminalKind.YulAutoKeyword:
                return this.rewriteYulAutoKeyword(node);
            case TerminalKind.YulBoolKeyword:
                return this.rewriteYulBoolKeyword(node);
            case TerminalKind.YulBreakKeyword:
                return this.rewriteYulBreakKeyword(node);
            case TerminalKind.YulBytesKeyword:
                return this.rewriteYulBytesKeyword(node);
            case TerminalKind.YulCallDataKeyword:
                return this.rewriteYulCallDataKeyword(node);
            case TerminalKind.YulCaseKeyword:
                return this.rewriteYulCaseKeyword(node);
            case TerminalKind.YulCatchKeyword:
                return this.rewriteYulCatchKeyword(node);
            case TerminalKind.YulConstantKeyword:
                return this.rewriteYulConstantKeyword(node);
            case TerminalKind.YulConstructorKeyword:
                return this.rewriteYulConstructorKeyword(node);
            case TerminalKind.YulContinueKeyword:
                return this.rewriteYulContinueKeyword(node);
            case TerminalKind.YulContractKeyword:
                return this.rewriteYulContractKeyword(node);
            case TerminalKind.YulCopyOfKeyword:
                return this.rewriteYulCopyOfKeyword(node);
            case TerminalKind.YulDaysKeyword:
                return this.rewriteYulDaysKeyword(node);
            case TerminalKind.YulDecimalLiteral:
                return this.rewriteYulDecimalLiteral(node);
            case TerminalKind.YulDefaultKeyword:
                return this.rewriteYulDefaultKeyword(node);
            case TerminalKind.YulDefineKeyword:
                return this.rewriteYulDefineKeyword(node);
            case TerminalKind.YulDeleteKeyword:
                return this.rewriteYulDeleteKeyword(node);
            case TerminalKind.YulDoKeyword:
                return this.rewriteYulDoKeyword(node);
            case TerminalKind.YulElseKeyword:
                return this.rewriteYulElseKeyword(node);
            case TerminalKind.YulEmitKeyword:
                return this.rewriteYulEmitKeyword(node);
            case TerminalKind.YulEnumKeyword:
                return this.rewriteYulEnumKeyword(node);
            case TerminalKind.YulEtherKeyword:
                return this.rewriteYulEtherKeyword(node);
            case TerminalKind.YulEventKeyword:
                return this.rewriteYulEventKeyword(node);
            case TerminalKind.YulExternalKeyword:
                return this.rewriteYulExternalKeyword(node);
            case TerminalKind.YulFallbackKeyword:
                return this.rewriteYulFallbackKeyword(node);
            case TerminalKind.YulFalseKeyword:
                return this.rewriteYulFalseKeyword(node);
            case TerminalKind.YulFinalKeyword:
                return this.rewriteYulFinalKeyword(node);
            case TerminalKind.YulFinneyKeyword:
                return this.rewriteYulFinneyKeyword(node);
            case TerminalKind.YulFixedKeyword:
                return this.rewriteYulFixedKeyword(node);
            case TerminalKind.YulForKeyword:
                return this.rewriteYulForKeyword(node);
            case TerminalKind.YulFunctionKeyword:
                return this.rewriteYulFunctionKeyword(node);
            case TerminalKind.YulGweiKeyword:
                return this.rewriteYulGweiKeyword(node);
            case TerminalKind.YulHexKeyword:
                return this.rewriteYulHexKeyword(node);
            case TerminalKind.YulHexLiteral:
                return this.rewriteYulHexLiteral(node);
            case TerminalKind.YulHoursKeyword:
                return this.rewriteYulHoursKeyword(node);
            case TerminalKind.YulIdentifier:
                return this.rewriteYulIdentifier(node);
            case TerminalKind.YulIfKeyword:
                return this.rewriteYulIfKeyword(node);
            case TerminalKind.YulImmutableKeyword:
                return this.rewriteYulImmutableKeyword(node);
            case TerminalKind.YulImplementsKeyword:
                return this.rewriteYulImplementsKeyword(node);
            case TerminalKind.YulImportKeyword:
                return this.rewriteYulImportKeyword(node);
            case TerminalKind.YulInKeyword:
                return this.rewriteYulInKeyword(node);
            case TerminalKind.YulIndexedKeyword:
                return this.rewriteYulIndexedKeyword(node);
            case TerminalKind.YulInlineKeyword:
                return this.rewriteYulInlineKeyword(node);
            case TerminalKind.YulIntKeyword:
                return this.rewriteYulIntKeyword(node);
            case TerminalKind.YulInterfaceKeyword:
                return this.rewriteYulInterfaceKeyword(node);
            case TerminalKind.YulInternalKeyword:
                return this.rewriteYulInternalKeyword(node);
            case TerminalKind.YulIsKeyword:
                return this.rewriteYulIsKeyword(node);
            case TerminalKind.YulLeaveKeyword:
                return this.rewriteYulLeaveKeyword(node);
            case TerminalKind.YulLetKeyword:
                return this.rewriteYulLetKeyword(node);
            case TerminalKind.YulLibraryKeyword:
                return this.rewriteYulLibraryKeyword(node);
            case TerminalKind.YulMacroKeyword:
                return this.rewriteYulMacroKeyword(node);
            case TerminalKind.YulMappingKeyword:
                return this.rewriteYulMappingKeyword(node);
            case TerminalKind.YulMatchKeyword:
                return this.rewriteYulMatchKeyword(node);
            case TerminalKind.YulMemoryKeyword:
                return this.rewriteYulMemoryKeyword(node);
            case TerminalKind.YulMinutesKeyword:
                return this.rewriteYulMinutesKeyword(node);
            case TerminalKind.YulModifierKeyword:
                return this.rewriteYulModifierKeyword(node);
            case TerminalKind.YulMutableKeyword:
                return this.rewriteYulMutableKeyword(node);
            case TerminalKind.YulNewKeyword:
                return this.rewriteYulNewKeyword(node);
            case TerminalKind.YulNullKeyword:
                return this.rewriteYulNullKeyword(node);
            case TerminalKind.YulOfKeyword:
                return this.rewriteYulOfKeyword(node);
            case TerminalKind.YulOverrideKeyword:
                return this.rewriteYulOverrideKeyword(node);
            case TerminalKind.YulPartialKeyword:
                return this.rewriteYulPartialKeyword(node);
            case TerminalKind.YulPayableKeyword:
                return this.rewriteYulPayableKeyword(node);
            case TerminalKind.YulPragmaKeyword:
                return this.rewriteYulPragmaKeyword(node);
            case TerminalKind.YulPrivateKeyword:
                return this.rewriteYulPrivateKeyword(node);
            case TerminalKind.YulPromiseKeyword:
                return this.rewriteYulPromiseKeyword(node);
            case TerminalKind.YulPublicKeyword:
                return this.rewriteYulPublicKeyword(node);
            case TerminalKind.YulPureKeyword:
                return this.rewriteYulPureKeyword(node);
            case TerminalKind.YulReceiveKeyword:
                return this.rewriteYulReceiveKeyword(node);
            case TerminalKind.YulReferenceKeyword:
                return this.rewriteYulReferenceKeyword(node);
            case TerminalKind.YulRelocatableKeyword:
                return this.rewriteYulRelocatableKeyword(node);
            case TerminalKind.YulReturnsKeyword:
                return this.rewriteYulReturnsKeyword(node);
            case TerminalKind.YulSealedKeyword:
                return this.rewriteYulSealedKeyword(node);
            case TerminalKind.YulSecondsKeyword:
                return this.rewriteYulSecondsKeyword(node);
            case TerminalKind.YulSizeOfKeyword:
                return this.rewriteYulSizeOfKeyword(node);
            case TerminalKind.YulStaticKeyword:
                return this.rewriteYulStaticKeyword(node);
            case TerminalKind.YulStorageKeyword:
                return this.rewriteYulStorageKeyword(node);
            case TerminalKind.YulStringKeyword:
                return this.rewriteYulStringKeyword(node);
            case TerminalKind.YulStructKeyword:
                return this.rewriteYulStructKeyword(node);
            case TerminalKind.YulSuperKeyword:
                return this.rewriteYulSuperKeyword(node);
            case TerminalKind.YulSupportsKeyword:
                return this.rewriteYulSupportsKeyword(node);
            case TerminalKind.YulSwitchKeyword:
                return this.rewriteYulSwitchKeyword(node);
            case TerminalKind.YulSzaboKeyword:
                return this.rewriteYulSzaboKeyword(node);
            case TerminalKind.YulThisKeyword:
                return this.rewriteYulThisKeyword(node);
            case TerminalKind.YulThrowKeyword:
                return this.rewriteYulThrowKeyword(node);
            case TerminalKind.YulTrueKeyword:
                return this.rewriteYulTrueKeyword(node);
            case TerminalKind.YulTryKeyword:
                return this.rewriteYulTryKeyword(node);
            case TerminalKind.YulTypeDefKeyword:
                return this.rewriteYulTypeDefKeyword(node);
            case TerminalKind.YulTypeKeyword:
                return this.rewriteYulTypeKeyword(node);
            case TerminalKind.YulTypeOfKeyword:
                return this.rewriteYulTypeOfKeyword(node);
            case TerminalKind.YulUfixedKeyword:
                return this.rewriteYulUfixedKeyword(node);
            case TerminalKind.YulUintKeyword:
                return this.rewriteYulUintKeyword(node);
            case TerminalKind.YulUncheckedKeyword:
                return this.rewriteYulUncheckedKeyword(node);
            case TerminalKind.YulUsingKeyword:
                return this.rewriteYulUsingKeyword(node);
            case TerminalKind.YulVarKeyword:
                return this.rewriteYulVarKeyword(node);
            case TerminalKind.YulViewKeyword:
                return this.rewriteYulViewKeyword(node);
            case TerminalKind.YulVirtualKeyword:
                return this.rewriteYulVirtualKeyword(node);
            case TerminalKind.YulWeeksKeyword:
                return this.rewriteYulWeeksKeyword(node);
            case TerminalKind.YulWeiKeyword:
                return this.rewriteYulWeiKeyword(node);
            case TerminalKind.YulWhileKeyword:
                return this.rewriteYulWhileKeyword(node);
            case TerminalKind.YulYearsKeyword:
                return this.rewriteYulYearsKeyword(node);
            case TerminalKind.Unrecognized:
                return this.rewriteUnrecognized(node);
            case TerminalKind.Missing:
                return this.rewriteMissing(node);
        }
    }
    /** @virtual Rewrites a `AbicoderPragma` node, recursively traversing the children (unless overriden). */
    rewriteAbicoderPragma(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `AbicoderVersion` node, recursively traversing the children (unless overriden). */
    rewriteAbicoderVersion(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `AdditiveExpression` node, recursively traversing the children (unless overriden). */
    rewriteAdditiveExpression(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `AddressType` node, recursively traversing the children (unless overriden). */
    rewriteAddressType(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `AndExpression` node, recursively traversing the children (unless overriden). */
    rewriteAndExpression(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `ArgumentsDeclaration` node, recursively traversing the children (unless overriden). */
    rewriteArgumentsDeclaration(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `ArrayExpression` node, recursively traversing the children (unless overriden). */
    rewriteArrayExpression(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `ArrayTypeName` node, recursively traversing the children (unless overriden). */
    rewriteArrayTypeName(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `ArrayValues` node, recursively traversing the children (unless overriden). */
    rewriteArrayValues(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `AssemblyFlags` node, recursively traversing the children (unless overriden). */
    rewriteAssemblyFlags(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `AssemblyFlagsDeclaration` node, recursively traversing the children (unless overriden). */
    rewriteAssemblyFlagsDeclaration(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `AssemblyStatement` node, recursively traversing the children (unless overriden). */
    rewriteAssemblyStatement(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `AssignmentExpression` node, recursively traversing the children (unless overriden). */
    rewriteAssignmentExpression(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `BitwiseAndExpression` node, recursively traversing the children (unless overriden). */
    rewriteBitwiseAndExpression(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `BitwiseOrExpression` node, recursively traversing the children (unless overriden). */
    rewriteBitwiseOrExpression(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `BitwiseXorExpression` node, recursively traversing the children (unless overriden). */
    rewriteBitwiseXorExpression(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `Block` node, recursively traversing the children (unless overriden). */
    rewriteBlock(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `BreakStatement` node, recursively traversing the children (unless overriden). */
    rewriteBreakStatement(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `CallOptions` node, recursively traversing the children (unless overriden). */
    rewriteCallOptions(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `CallOptionsExpression` node, recursively traversing the children (unless overriden). */
    rewriteCallOptionsExpression(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `CatchClause` node, recursively traversing the children (unless overriden). */
    rewriteCatchClause(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `CatchClauseError` node, recursively traversing the children (unless overriden). */
    rewriteCatchClauseError(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `CatchClauses` node, recursively traversing the children (unless overriden). */
    rewriteCatchClauses(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `ConditionalExpression` node, recursively traversing the children (unless overriden). */
    rewriteConditionalExpression(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `ConstantDefinition` node, recursively traversing the children (unless overriden). */
    rewriteConstantDefinition(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `ConstructorAttribute` node, recursively traversing the children (unless overriden). */
    rewriteConstructorAttribute(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `ConstructorAttributes` node, recursively traversing the children (unless overriden). */
    rewriteConstructorAttributes(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `ConstructorDefinition` node, recursively traversing the children (unless overriden). */
    rewriteConstructorDefinition(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `ContinueStatement` node, recursively traversing the children (unless overriden). */
    rewriteContinueStatement(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `ContractDefinition` node, recursively traversing the children (unless overriden). */
    rewriteContractDefinition(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `ContractMember` node, recursively traversing the children (unless overriden). */
    rewriteContractMember(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `ContractMembers` node, recursively traversing the children (unless overriden). */
    rewriteContractMembers(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `ContractSpecifier` node, recursively traversing the children (unless overriden). */
    rewriteContractSpecifier(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `ContractSpecifiers` node, recursively traversing the children (unless overriden). */
    rewriteContractSpecifiers(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `DecimalNumberExpression` node, recursively traversing the children (unless overriden). */
    rewriteDecimalNumberExpression(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `DoWhileStatement` node, recursively traversing the children (unless overriden). */
    rewriteDoWhileStatement(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `ElementaryType` node, recursively traversing the children (unless overriden). */
    rewriteElementaryType(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `ElseBranch` node, recursively traversing the children (unless overriden). */
    rewriteElseBranch(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `EmitStatement` node, recursively traversing the children (unless overriden). */
    rewriteEmitStatement(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `EnumDefinition` node, recursively traversing the children (unless overriden). */
    rewriteEnumDefinition(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `EnumMembers` node, recursively traversing the children (unless overriden). */
    rewriteEnumMembers(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `EqualityExpression` node, recursively traversing the children (unless overriden). */
    rewriteEqualityExpression(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `ErrorDefinition` node, recursively traversing the children (unless overriden). */
    rewriteErrorDefinition(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `ErrorParameter` node, recursively traversing the children (unless overriden). */
    rewriteErrorParameter(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `ErrorParameters` node, recursively traversing the children (unless overriden). */
    rewriteErrorParameters(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `ErrorParametersDeclaration` node, recursively traversing the children (unless overriden). */
    rewriteErrorParametersDeclaration(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `EventDefinition` node, recursively traversing the children (unless overriden). */
    rewriteEventDefinition(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `EventParameter` node, recursively traversing the children (unless overriden). */
    rewriteEventParameter(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `EventParameters` node, recursively traversing the children (unless overriden). */
    rewriteEventParameters(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `EventParametersDeclaration` node, recursively traversing the children (unless overriden). */
    rewriteEventParametersDeclaration(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `ExperimentalFeature` node, recursively traversing the children (unless overriden). */
    rewriteExperimentalFeature(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `ExperimentalPragma` node, recursively traversing the children (unless overriden). */
    rewriteExperimentalPragma(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `ExponentiationExpression` node, recursively traversing the children (unless overriden). */
    rewriteExponentiationExpression(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `Expression` node, recursively traversing the children (unless overriden). */
    rewriteExpression(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `ExpressionStatement` node, recursively traversing the children (unless overriden). */
    rewriteExpressionStatement(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `FallbackFunctionAttribute` node, recursively traversing the children (unless overriden). */
    rewriteFallbackFunctionAttribute(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `FallbackFunctionAttributes` node, recursively traversing the children (unless overriden). */
    rewriteFallbackFunctionAttributes(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `FallbackFunctionDefinition` node, recursively traversing the children (unless overriden). */
    rewriteFallbackFunctionDefinition(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `ForStatement` node, recursively traversing the children (unless overriden). */
    rewriteForStatement(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `ForStatementCondition` node, recursively traversing the children (unless overriden). */
    rewriteForStatementCondition(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `ForStatementInitialization` node, recursively traversing the children (unless overriden). */
    rewriteForStatementInitialization(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `FunctionAttribute` node, recursively traversing the children (unless overriden). */
    rewriteFunctionAttribute(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `FunctionAttributes` node, recursively traversing the children (unless overriden). */
    rewriteFunctionAttributes(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `FunctionBody` node, recursively traversing the children (unless overriden). */
    rewriteFunctionBody(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `FunctionCallExpression` node, recursively traversing the children (unless overriden). */
    rewriteFunctionCallExpression(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `FunctionDefinition` node, recursively traversing the children (unless overriden). */
    rewriteFunctionDefinition(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `FunctionName` node, recursively traversing the children (unless overriden). */
    rewriteFunctionName(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `FunctionType` node, recursively traversing the children (unless overriden). */
    rewriteFunctionType(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `FunctionTypeAttribute` node, recursively traversing the children (unless overriden). */
    rewriteFunctionTypeAttribute(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `FunctionTypeAttributes` node, recursively traversing the children (unless overriden). */
    rewriteFunctionTypeAttributes(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `HexNumberExpression` node, recursively traversing the children (unless overriden). */
    rewriteHexNumberExpression(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `HexStringLiteral` node, recursively traversing the children (unless overriden). */
    rewriteHexStringLiteral(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `HexStringLiterals` node, recursively traversing the children (unless overriden). */
    rewriteHexStringLiterals(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `IdentifierPath` node, recursively traversing the children (unless overriden). */
    rewriteIdentifierPath(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `IfStatement` node, recursively traversing the children (unless overriden). */
    rewriteIfStatement(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `ImportAlias` node, recursively traversing the children (unless overriden). */
    rewriteImportAlias(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `ImportClause` node, recursively traversing the children (unless overriden). */
    rewriteImportClause(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `ImportDeconstruction` node, recursively traversing the children (unless overriden). */
    rewriteImportDeconstruction(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `ImportDeconstructionSymbol` node, recursively traversing the children (unless overriden). */
    rewriteImportDeconstructionSymbol(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `ImportDeconstructionSymbols` node, recursively traversing the children (unless overriden). */
    rewriteImportDeconstructionSymbols(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `ImportDirective` node, recursively traversing the children (unless overriden). */
    rewriteImportDirective(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `IndexAccessEnd` node, recursively traversing the children (unless overriden). */
    rewriteIndexAccessEnd(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `IndexAccessExpression` node, recursively traversing the children (unless overriden). */
    rewriteIndexAccessExpression(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `InequalityExpression` node, recursively traversing the children (unless overriden). */
    rewriteInequalityExpression(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `InheritanceSpecifier` node, recursively traversing the children (unless overriden). */
    rewriteInheritanceSpecifier(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `InheritanceType` node, recursively traversing the children (unless overriden). */
    rewriteInheritanceType(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `InheritanceTypes` node, recursively traversing the children (unless overriden). */
    rewriteInheritanceTypes(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `InterfaceDefinition` node, recursively traversing the children (unless overriden). */
    rewriteInterfaceDefinition(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `InterfaceMembers` node, recursively traversing the children (unless overriden). */
    rewriteInterfaceMembers(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `LibraryDefinition` node, recursively traversing the children (unless overriden). */
    rewriteLibraryDefinition(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `LibraryMembers` node, recursively traversing the children (unless overriden). */
    rewriteLibraryMembers(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `MappingKey` node, recursively traversing the children (unless overriden). */
    rewriteMappingKey(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `MappingKeyType` node, recursively traversing the children (unless overriden). */
    rewriteMappingKeyType(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `MappingType` node, recursively traversing the children (unless overriden). */
    rewriteMappingType(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `MappingValue` node, recursively traversing the children (unless overriden). */
    rewriteMappingValue(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `MemberAccessExpression` node, recursively traversing the children (unless overriden). */
    rewriteMemberAccessExpression(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `ModifierAttribute` node, recursively traversing the children (unless overriden). */
    rewriteModifierAttribute(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `ModifierAttributes` node, recursively traversing the children (unless overriden). */
    rewriteModifierAttributes(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `ModifierDefinition` node, recursively traversing the children (unless overriden). */
    rewriteModifierDefinition(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `ModifierInvocation` node, recursively traversing the children (unless overriden). */
    rewriteModifierInvocation(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `MultiplicativeExpression` node, recursively traversing the children (unless overriden). */
    rewriteMultiplicativeExpression(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `NamedArgument` node, recursively traversing the children (unless overriden). */
    rewriteNamedArgument(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `NamedArgumentGroup` node, recursively traversing the children (unless overriden). */
    rewriteNamedArgumentGroup(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `NamedArguments` node, recursively traversing the children (unless overriden). */
    rewriteNamedArguments(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `NamedArgumentsDeclaration` node, recursively traversing the children (unless overriden). */
    rewriteNamedArgumentsDeclaration(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `NamedImport` node, recursively traversing the children (unless overriden). */
    rewriteNamedImport(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `NewExpression` node, recursively traversing the children (unless overriden). */
    rewriteNewExpression(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `NumberUnit` node, recursively traversing the children (unless overriden). */
    rewriteNumberUnit(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `OrExpression` node, recursively traversing the children (unless overriden). */
    rewriteOrExpression(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `OverridePaths` node, recursively traversing the children (unless overriden). */
    rewriteOverridePaths(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `OverridePathsDeclaration` node, recursively traversing the children (unless overriden). */
    rewriteOverridePathsDeclaration(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `OverrideSpecifier` node, recursively traversing the children (unless overriden). */
    rewriteOverrideSpecifier(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `Parameter` node, recursively traversing the children (unless overriden). */
    rewriteParameter(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `Parameters` node, recursively traversing the children (unless overriden). */
    rewriteParameters(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `ParametersDeclaration` node, recursively traversing the children (unless overriden). */
    rewriteParametersDeclaration(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `PathImport` node, recursively traversing the children (unless overriden). */
    rewritePathImport(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `PositionalArguments` node, recursively traversing the children (unless overriden). */
    rewritePositionalArguments(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `PositionalArgumentsDeclaration` node, recursively traversing the children (unless overriden). */
    rewritePositionalArgumentsDeclaration(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `PostfixExpression` node, recursively traversing the children (unless overriden). */
    rewritePostfixExpression(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `Pragma` node, recursively traversing the children (unless overriden). */
    rewritePragma(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `PragmaDirective` node, recursively traversing the children (unless overriden). */
    rewritePragmaDirective(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `PrefixExpression` node, recursively traversing the children (unless overriden). */
    rewritePrefixExpression(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `ReceiveFunctionAttribute` node, recursively traversing the children (unless overriden). */
    rewriteReceiveFunctionAttribute(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `ReceiveFunctionAttributes` node, recursively traversing the children (unless overriden). */
    rewriteReceiveFunctionAttributes(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `ReceiveFunctionDefinition` node, recursively traversing the children (unless overriden). */
    rewriteReceiveFunctionDefinition(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `ReturnStatement` node, recursively traversing the children (unless overriden). */
    rewriteReturnStatement(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `ReturnsDeclaration` node, recursively traversing the children (unless overriden). */
    rewriteReturnsDeclaration(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `RevertStatement` node, recursively traversing the children (unless overriden). */
    rewriteRevertStatement(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `ShiftExpression` node, recursively traversing the children (unless overriden). */
    rewriteShiftExpression(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `SimpleVersionLiteral` node, recursively traversing the children (unless overriden). */
    rewriteSimpleVersionLiteral(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `SourceUnit` node, recursively traversing the children (unless overriden). */
    rewriteSourceUnit(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `SourceUnitMember` node, recursively traversing the children (unless overriden). */
    rewriteSourceUnitMember(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `SourceUnitMembers` node, recursively traversing the children (unless overriden). */
    rewriteSourceUnitMembers(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `StateVariableAttribute` node, recursively traversing the children (unless overriden). */
    rewriteStateVariableAttribute(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `StateVariableAttributes` node, recursively traversing the children (unless overriden). */
    rewriteStateVariableAttributes(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `StateVariableDefinition` node, recursively traversing the children (unless overriden). */
    rewriteStateVariableDefinition(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `StateVariableDefinitionValue` node, recursively traversing the children (unless overriden). */
    rewriteStateVariableDefinitionValue(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `Statement` node, recursively traversing the children (unless overriden). */
    rewriteStatement(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `Statements` node, recursively traversing the children (unless overriden). */
    rewriteStatements(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `StorageLayoutSpecifier` node, recursively traversing the children (unless overriden). */
    rewriteStorageLayoutSpecifier(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `StorageLocation` node, recursively traversing the children (unless overriden). */
    rewriteStorageLocation(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `StringExpression` node, recursively traversing the children (unless overriden). */
    rewriteStringExpression(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `StringLiteral` node, recursively traversing the children (unless overriden). */
    rewriteStringLiteral(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `StringLiterals` node, recursively traversing the children (unless overriden). */
    rewriteStringLiterals(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `StructDefinition` node, recursively traversing the children (unless overriden). */
    rewriteStructDefinition(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `StructMember` node, recursively traversing the children (unless overriden). */
    rewriteStructMember(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `StructMembers` node, recursively traversing the children (unless overriden). */
    rewriteStructMembers(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `ThrowStatement` node, recursively traversing the children (unless overriden). */
    rewriteThrowStatement(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `TryStatement` node, recursively traversing the children (unless overriden). */
    rewriteTryStatement(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `TupleDeconstructionElement` node, recursively traversing the children (unless overriden). */
    rewriteTupleDeconstructionElement(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `TupleDeconstructionElements` node, recursively traversing the children (unless overriden). */
    rewriteTupleDeconstructionElements(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `TupleDeconstructionStatement` node, recursively traversing the children (unless overriden). */
    rewriteTupleDeconstructionStatement(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `TupleExpression` node, recursively traversing the children (unless overriden). */
    rewriteTupleExpression(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `TupleMember` node, recursively traversing the children (unless overriden). */
    rewriteTupleMember(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `TupleValue` node, recursively traversing the children (unless overriden). */
    rewriteTupleValue(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `TupleValues` node, recursively traversing the children (unless overriden). */
    rewriteTupleValues(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `TypeExpression` node, recursively traversing the children (unless overriden). */
    rewriteTypeExpression(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `TypeName` node, recursively traversing the children (unless overriden). */
    rewriteTypeName(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `TypedTupleMember` node, recursively traversing the children (unless overriden). */
    rewriteTypedTupleMember(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `UncheckedBlock` node, recursively traversing the children (unless overriden). */
    rewriteUncheckedBlock(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `UnicodeStringLiteral` node, recursively traversing the children (unless overriden). */
    rewriteUnicodeStringLiteral(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `UnicodeStringLiterals` node, recursively traversing the children (unless overriden). */
    rewriteUnicodeStringLiterals(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `UnnamedFunctionAttribute` node, recursively traversing the children (unless overriden). */
    rewriteUnnamedFunctionAttribute(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `UnnamedFunctionAttributes` node, recursively traversing the children (unless overriden). */
    rewriteUnnamedFunctionAttributes(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `UnnamedFunctionDefinition` node, recursively traversing the children (unless overriden). */
    rewriteUnnamedFunctionDefinition(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `UntypedTupleMember` node, recursively traversing the children (unless overriden). */
    rewriteUntypedTupleMember(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `UserDefinedValueTypeDefinition` node, recursively traversing the children (unless overriden). */
    rewriteUserDefinedValueTypeDefinition(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `UsingAlias` node, recursively traversing the children (unless overriden). */
    rewriteUsingAlias(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `UsingClause` node, recursively traversing the children (unless overriden). */
    rewriteUsingClause(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `UsingDeconstruction` node, recursively traversing the children (unless overriden). */
    rewriteUsingDeconstruction(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `UsingDeconstructionSymbol` node, recursively traversing the children (unless overriden). */
    rewriteUsingDeconstructionSymbol(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `UsingDeconstructionSymbols` node, recursively traversing the children (unless overriden). */
    rewriteUsingDeconstructionSymbols(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `UsingDirective` node, recursively traversing the children (unless overriden). */
    rewriteUsingDirective(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `UsingOperator` node, recursively traversing the children (unless overriden). */
    rewriteUsingOperator(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `UsingTarget` node, recursively traversing the children (unless overriden). */
    rewriteUsingTarget(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `VariableDeclarationStatement` node, recursively traversing the children (unless overriden). */
    rewriteVariableDeclarationStatement(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `VariableDeclarationType` node, recursively traversing the children (unless overriden). */
    rewriteVariableDeclarationType(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `VariableDeclarationValue` node, recursively traversing the children (unless overriden). */
    rewriteVariableDeclarationValue(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `VersionExpression` node, recursively traversing the children (unless overriden). */
    rewriteVersionExpression(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `VersionExpressionSet` node, recursively traversing the children (unless overriden). */
    rewriteVersionExpressionSet(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `VersionExpressionSets` node, recursively traversing the children (unless overriden). */
    rewriteVersionExpressionSets(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `VersionLiteral` node, recursively traversing the children (unless overriden). */
    rewriteVersionLiteral(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `VersionOperator` node, recursively traversing the children (unless overriden). */
    rewriteVersionOperator(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `VersionPragma` node, recursively traversing the children (unless overriden). */
    rewriteVersionPragma(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `VersionRange` node, recursively traversing the children (unless overriden). */
    rewriteVersionRange(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `VersionTerm` node, recursively traversing the children (unless overriden). */
    rewriteVersionTerm(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `WhileStatement` node, recursively traversing the children (unless overriden). */
    rewriteWhileStatement(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `YulArguments` node, recursively traversing the children (unless overriden). */
    rewriteYulArguments(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `YulAssignmentOperator` node, recursively traversing the children (unless overriden). */
    rewriteYulAssignmentOperator(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `YulBlock` node, recursively traversing the children (unless overriden). */
    rewriteYulBlock(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `YulBreakStatement` node, recursively traversing the children (unless overriden). */
    rewriteYulBreakStatement(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `YulColonAndEqual` node, recursively traversing the children (unless overriden). */
    rewriteYulColonAndEqual(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `YulContinueStatement` node, recursively traversing the children (unless overriden). */
    rewriteYulContinueStatement(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `YulDefaultCase` node, recursively traversing the children (unless overriden). */
    rewriteYulDefaultCase(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `YulEqualAndColon` node, recursively traversing the children (unless overriden). */
    rewriteYulEqualAndColon(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `YulExpression` node, recursively traversing the children (unless overriden). */
    rewriteYulExpression(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `YulForStatement` node, recursively traversing the children (unless overriden). */
    rewriteYulForStatement(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `YulFunctionCallExpression` node, recursively traversing the children (unless overriden). */
    rewriteYulFunctionCallExpression(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `YulFunctionDefinition` node, recursively traversing the children (unless overriden). */
    rewriteYulFunctionDefinition(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `YulIfStatement` node, recursively traversing the children (unless overriden). */
    rewriteYulIfStatement(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `YulLabel` node, recursively traversing the children (unless overriden). */
    rewriteYulLabel(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `YulLeaveStatement` node, recursively traversing the children (unless overriden). */
    rewriteYulLeaveStatement(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `YulLiteral` node, recursively traversing the children (unless overriden). */
    rewriteYulLiteral(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `YulParameters` node, recursively traversing the children (unless overriden). */
    rewriteYulParameters(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `YulParametersDeclaration` node, recursively traversing the children (unless overriden). */
    rewriteYulParametersDeclaration(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `YulPath` node, recursively traversing the children (unless overriden). */
    rewriteYulPath(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `YulPaths` node, recursively traversing the children (unless overriden). */
    rewriteYulPaths(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `YulReturnsDeclaration` node, recursively traversing the children (unless overriden). */
    rewriteYulReturnsDeclaration(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `YulStackAssignmentOperator` node, recursively traversing the children (unless overriden). */
    rewriteYulStackAssignmentOperator(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `YulStackAssignmentStatement` node, recursively traversing the children (unless overriden). */
    rewriteYulStackAssignmentStatement(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `YulStatement` node, recursively traversing the children (unless overriden). */
    rewriteYulStatement(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `YulStatements` node, recursively traversing the children (unless overriden). */
    rewriteYulStatements(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `YulSwitchCase` node, recursively traversing the children (unless overriden). */
    rewriteYulSwitchCase(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `YulSwitchCases` node, recursively traversing the children (unless overriden). */
    rewriteYulSwitchCases(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `YulSwitchStatement` node, recursively traversing the children (unless overriden). */
    rewriteYulSwitchStatement(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `YulValueCase` node, recursively traversing the children (unless overriden). */
    rewriteYulValueCase(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `YulVariableAssignmentStatement` node, recursively traversing the children (unless overriden). */
    rewriteYulVariableAssignmentStatement(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `YulVariableDeclarationStatement` node, recursively traversing the children (unless overriden). */
    rewriteYulVariableDeclarationStatement(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `YulVariableDeclarationValue` node, recursively traversing the children (unless overriden). */
    rewriteYulVariableDeclarationValue(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `YulVariableNames` node, recursively traversing the children (unless overriden). */
    rewriteYulVariableNames(node) {
        return this.rewriteChildren(node);
    }
    /** @virtual Rewrites a `ABIEncoderV2Keyword` node. */
    rewriteABIEncoderV2Keyword(node) {
        return node;
    }
    /** @virtual Rewrites a `AbicoderKeyword` node. */
    rewriteAbicoderKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `AbicoderV1Keyword` node. */
    rewriteAbicoderV1Keyword(node) {
        return node;
    }
    /** @virtual Rewrites a `AbicoderV2Keyword` node. */
    rewriteAbicoderV2Keyword(node) {
        return node;
    }
    /** @virtual Rewrites a `AbstractKeyword` node. */
    rewriteAbstractKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `AddressKeyword` node. */
    rewriteAddressKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `AfterKeyword` node. */
    rewriteAfterKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `AliasKeyword` node. */
    rewriteAliasKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `Ampersand` node. */
    rewriteAmpersand(node) {
        return node;
    }
    /** @virtual Rewrites a `AmpersandAmpersand` node. */
    rewriteAmpersandAmpersand(node) {
        return node;
    }
    /** @virtual Rewrites a `AmpersandEqual` node. */
    rewriteAmpersandEqual(node) {
        return node;
    }
    /** @virtual Rewrites a `AnonymousKeyword` node. */
    rewriteAnonymousKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `ApplyKeyword` node. */
    rewriteApplyKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `AsKeyword` node. */
    rewriteAsKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `AssemblyKeyword` node. */
    rewriteAssemblyKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `Asterisk` node. */
    rewriteAsterisk(node) {
        return node;
    }
    /** @virtual Rewrites a `AsteriskAsterisk` node. */
    rewriteAsteriskAsterisk(node) {
        return node;
    }
    /** @virtual Rewrites a `AsteriskEqual` node. */
    rewriteAsteriskEqual(node) {
        return node;
    }
    /** @virtual Rewrites a `AtKeyword` node. */
    rewriteAtKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `AutoKeyword` node. */
    rewriteAutoKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `Bang` node. */
    rewriteBang(node) {
        return node;
    }
    /** @virtual Rewrites a `BangEqual` node. */
    rewriteBangEqual(node) {
        return node;
    }
    /** @virtual Rewrites a `Bar` node. */
    rewriteBar(node) {
        return node;
    }
    /** @virtual Rewrites a `BarBar` node. */
    rewriteBarBar(node) {
        return node;
    }
    /** @virtual Rewrites a `BarEqual` node. */
    rewriteBarEqual(node) {
        return node;
    }
    /** @virtual Rewrites a `BoolKeyword` node. */
    rewriteBoolKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `BreakKeyword` node. */
    rewriteBreakKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `ByteKeyword` node. */
    rewriteByteKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `BytesKeyword` node. */
    rewriteBytesKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `CallDataKeyword` node. */
    rewriteCallDataKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `Caret` node. */
    rewriteCaret(node) {
        return node;
    }
    /** @virtual Rewrites a `CaretEqual` node. */
    rewriteCaretEqual(node) {
        return node;
    }
    /** @virtual Rewrites a `CaseKeyword` node. */
    rewriteCaseKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `CatchKeyword` node. */
    rewriteCatchKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `CloseBrace` node. */
    rewriteCloseBrace(node) {
        return node;
    }
    /** @virtual Rewrites a `CloseBracket` node. */
    rewriteCloseBracket(node) {
        return node;
    }
    /** @virtual Rewrites a `CloseParen` node. */
    rewriteCloseParen(node) {
        return node;
    }
    /** @virtual Rewrites a `Colon` node. */
    rewriteColon(node) {
        return node;
    }
    /** @virtual Rewrites a `ColonEqual` node. */
    rewriteColonEqual(node) {
        return node;
    }
    /** @virtual Rewrites a `Comma` node. */
    rewriteComma(node) {
        return node;
    }
    /** @virtual Rewrites a `ConstantKeyword` node. */
    rewriteConstantKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `ConstructorKeyword` node. */
    rewriteConstructorKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `ContinueKeyword` node. */
    rewriteContinueKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `ContractKeyword` node. */
    rewriteContractKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `CopyOfKeyword` node. */
    rewriteCopyOfKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `DaysKeyword` node. */
    rewriteDaysKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `DecimalLiteral` node. */
    rewriteDecimalLiteral(node) {
        return node;
    }
    /** @virtual Rewrites a `DefaultKeyword` node. */
    rewriteDefaultKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `DefineKeyword` node. */
    rewriteDefineKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `DeleteKeyword` node. */
    rewriteDeleteKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `DoKeyword` node. */
    rewriteDoKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `DoubleQuotedHexStringLiteral` node. */
    rewriteDoubleQuotedHexStringLiteral(node) {
        return node;
    }
    /** @virtual Rewrites a `DoubleQuotedStringLiteral` node. */
    rewriteDoubleQuotedStringLiteral(node) {
        return node;
    }
    /** @virtual Rewrites a `DoubleQuotedUnicodeStringLiteral` node. */
    rewriteDoubleQuotedUnicodeStringLiteral(node) {
        return node;
    }
    /** @virtual Rewrites a `DoubleQuotedVersionLiteral` node. */
    rewriteDoubleQuotedVersionLiteral(node) {
        return node;
    }
    /** @virtual Rewrites a `ElseKeyword` node. */
    rewriteElseKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `EmitKeyword` node. */
    rewriteEmitKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `EndOfLine` node. */
    rewriteEndOfLine(node) {
        return node;
    }
    /** @virtual Rewrites a `EnumKeyword` node. */
    rewriteEnumKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `Equal` node. */
    rewriteEqual(node) {
        return node;
    }
    /** @virtual Rewrites a `EqualColon` node. */
    rewriteEqualColon(node) {
        return node;
    }
    /** @virtual Rewrites a `EqualEqual` node. */
    rewriteEqualEqual(node) {
        return node;
    }
    /** @virtual Rewrites a `EqualGreaterThan` node. */
    rewriteEqualGreaterThan(node) {
        return node;
    }
    /** @virtual Rewrites a `ErrorKeyword` node. */
    rewriteErrorKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `EtherKeyword` node. */
    rewriteEtherKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `EventKeyword` node. */
    rewriteEventKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `ExperimentalKeyword` node. */
    rewriteExperimentalKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `ExternalKeyword` node. */
    rewriteExternalKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `FallbackKeyword` node. */
    rewriteFallbackKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `FalseKeyword` node. */
    rewriteFalseKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `FinalKeyword` node. */
    rewriteFinalKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `FinneyKeyword` node. */
    rewriteFinneyKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `FixedKeyword` node. */
    rewriteFixedKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `ForKeyword` node. */
    rewriteForKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `FromKeyword` node. */
    rewriteFromKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `FunctionKeyword` node. */
    rewriteFunctionKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `GlobalKeyword` node. */
    rewriteGlobalKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `GreaterThan` node. */
    rewriteGreaterThan(node) {
        return node;
    }
    /** @virtual Rewrites a `GreaterThanEqual` node. */
    rewriteGreaterThanEqual(node) {
        return node;
    }
    /** @virtual Rewrites a `GreaterThanGreaterThan` node. */
    rewriteGreaterThanGreaterThan(node) {
        return node;
    }
    /** @virtual Rewrites a `GreaterThanGreaterThanEqual` node. */
    rewriteGreaterThanGreaterThanEqual(node) {
        return node;
    }
    /** @virtual Rewrites a `GreaterThanGreaterThanGreaterThan` node. */
    rewriteGreaterThanGreaterThanGreaterThan(node) {
        return node;
    }
    /** @virtual Rewrites a `GreaterThanGreaterThanGreaterThanEqual` node. */
    rewriteGreaterThanGreaterThanGreaterThanEqual(node) {
        return node;
    }
    /** @virtual Rewrites a `GweiKeyword` node. */
    rewriteGweiKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `HexKeyword` node. */
    rewriteHexKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `HexLiteral` node. */
    rewriteHexLiteral(node) {
        return node;
    }
    /** @virtual Rewrites a `HoursKeyword` node. */
    rewriteHoursKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `Identifier` node. */
    rewriteIdentifier(node) {
        return node;
    }
    /** @virtual Rewrites a `IfKeyword` node. */
    rewriteIfKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `ImmutableKeyword` node. */
    rewriteImmutableKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `ImplementsKeyword` node. */
    rewriteImplementsKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `ImportKeyword` node. */
    rewriteImportKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `InKeyword` node. */
    rewriteInKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `IndexedKeyword` node. */
    rewriteIndexedKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `InlineKeyword` node. */
    rewriteInlineKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `IntKeyword` node. */
    rewriteIntKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `InterfaceKeyword` node. */
    rewriteInterfaceKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `InternalKeyword` node. */
    rewriteInternalKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `IsKeyword` node. */
    rewriteIsKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `LayoutKeyword` node. */
    rewriteLayoutKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `LessThan` node. */
    rewriteLessThan(node) {
        return node;
    }
    /** @virtual Rewrites a `LessThanEqual` node. */
    rewriteLessThanEqual(node) {
        return node;
    }
    /** @virtual Rewrites a `LessThanLessThan` node. */
    rewriteLessThanLessThan(node) {
        return node;
    }
    /** @virtual Rewrites a `LessThanLessThanEqual` node. */
    rewriteLessThanLessThanEqual(node) {
        return node;
    }
    /** @virtual Rewrites a `LetKeyword` node. */
    rewriteLetKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `LibraryKeyword` node. */
    rewriteLibraryKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `MacroKeyword` node. */
    rewriteMacroKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `MappingKeyword` node. */
    rewriteMappingKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `MatchKeyword` node. */
    rewriteMatchKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `MemoryKeyword` node. */
    rewriteMemoryKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `Minus` node. */
    rewriteMinus(node) {
        return node;
    }
    /** @virtual Rewrites a `MinusEqual` node. */
    rewriteMinusEqual(node) {
        return node;
    }
    /** @virtual Rewrites a `MinusGreaterThan` node. */
    rewriteMinusGreaterThan(node) {
        return node;
    }
    /** @virtual Rewrites a `MinusMinus` node. */
    rewriteMinusMinus(node) {
        return node;
    }
    /** @virtual Rewrites a `MinutesKeyword` node. */
    rewriteMinutesKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `ModifierKeyword` node. */
    rewriteModifierKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `MultiLineComment` node. */
    rewriteMultiLineComment(node) {
        return node;
    }
    /** @virtual Rewrites a `MultiLineNatSpecComment` node. */
    rewriteMultiLineNatSpecComment(node) {
        return node;
    }
    /** @virtual Rewrites a `MutableKeyword` node. */
    rewriteMutableKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `NewKeyword` node. */
    rewriteNewKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `NullKeyword` node. */
    rewriteNullKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `OfKeyword` node. */
    rewriteOfKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `OpenBrace` node. */
    rewriteOpenBrace(node) {
        return node;
    }
    /** @virtual Rewrites a `OpenBracket` node. */
    rewriteOpenBracket(node) {
        return node;
    }
    /** @virtual Rewrites a `OpenParen` node. */
    rewriteOpenParen(node) {
        return node;
    }
    /** @virtual Rewrites a `OverrideKeyword` node. */
    rewriteOverrideKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `PartialKeyword` node. */
    rewritePartialKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `PayableKeyword` node. */
    rewritePayableKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `Percent` node. */
    rewritePercent(node) {
        return node;
    }
    /** @virtual Rewrites a `PercentEqual` node. */
    rewritePercentEqual(node) {
        return node;
    }
    /** @virtual Rewrites a `Period` node. */
    rewritePeriod(node) {
        return node;
    }
    /** @virtual Rewrites a `Plus` node. */
    rewritePlus(node) {
        return node;
    }
    /** @virtual Rewrites a `PlusEqual` node. */
    rewritePlusEqual(node) {
        return node;
    }
    /** @virtual Rewrites a `PlusPlus` node. */
    rewritePlusPlus(node) {
        return node;
    }
    /** @virtual Rewrites a `PragmaKeyword` node. */
    rewritePragmaKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `PrivateKeyword` node. */
    rewritePrivateKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `PromiseKeyword` node. */
    rewritePromiseKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `PublicKeyword` node. */
    rewritePublicKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `PureKeyword` node. */
    rewritePureKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `QuestionMark` node. */
    rewriteQuestionMark(node) {
        return node;
    }
    /** @virtual Rewrites a `ReceiveKeyword` node. */
    rewriteReceiveKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `ReferenceKeyword` node. */
    rewriteReferenceKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `RelocatableKeyword` node. */
    rewriteRelocatableKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `ReturnKeyword` node. */
    rewriteReturnKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `ReturnsKeyword` node. */
    rewriteReturnsKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `RevertKeyword` node. */
    rewriteRevertKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `SMTCheckerKeyword` node. */
    rewriteSMTCheckerKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `SealedKeyword` node. */
    rewriteSealedKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `SecondsKeyword` node. */
    rewriteSecondsKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `Semicolon` node. */
    rewriteSemicolon(node) {
        return node;
    }
    /** @virtual Rewrites a `SingleLineComment` node. */
    rewriteSingleLineComment(node) {
        return node;
    }
    /** @virtual Rewrites a `SingleLineNatSpecComment` node. */
    rewriteSingleLineNatSpecComment(node) {
        return node;
    }
    /** @virtual Rewrites a `SingleQuotedHexStringLiteral` node. */
    rewriteSingleQuotedHexStringLiteral(node) {
        return node;
    }
    /** @virtual Rewrites a `SingleQuotedStringLiteral` node. */
    rewriteSingleQuotedStringLiteral(node) {
        return node;
    }
    /** @virtual Rewrites a `SingleQuotedUnicodeStringLiteral` node. */
    rewriteSingleQuotedUnicodeStringLiteral(node) {
        return node;
    }
    /** @virtual Rewrites a `SingleQuotedVersionLiteral` node. */
    rewriteSingleQuotedVersionLiteral(node) {
        return node;
    }
    /** @virtual Rewrites a `SizeOfKeyword` node. */
    rewriteSizeOfKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `Slash` node. */
    rewriteSlash(node) {
        return node;
    }
    /** @virtual Rewrites a `SlashEqual` node. */
    rewriteSlashEqual(node) {
        return node;
    }
    /** @virtual Rewrites a `SolidityKeyword` node. */
    rewriteSolidityKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `StaticKeyword` node. */
    rewriteStaticKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `StorageKeyword` node. */
    rewriteStorageKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `StringKeyword` node. */
    rewriteStringKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `StructKeyword` node. */
    rewriteStructKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `SuperKeyword` node. */
    rewriteSuperKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `SupportsKeyword` node. */
    rewriteSupportsKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `SwitchKeyword` node. */
    rewriteSwitchKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `SzaboKeyword` node. */
    rewriteSzaboKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `ThisKeyword` node. */
    rewriteThisKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `ThrowKeyword` node. */
    rewriteThrowKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `Tilde` node. */
    rewriteTilde(node) {
        return node;
    }
    /** @virtual Rewrites a `TransientKeyword` node. */
    rewriteTransientKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `TrueKeyword` node. */
    rewriteTrueKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `TryKeyword` node. */
    rewriteTryKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `TypeDefKeyword` node. */
    rewriteTypeDefKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `TypeKeyword` node. */
    rewriteTypeKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `TypeOfKeyword` node. */
    rewriteTypeOfKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `UfixedKeyword` node. */
    rewriteUfixedKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `UintKeyword` node. */
    rewriteUintKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `UncheckedKeyword` node. */
    rewriteUncheckedKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `UsingKeyword` node. */
    rewriteUsingKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `VarKeyword` node. */
    rewriteVarKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `VersionSpecifier` node. */
    rewriteVersionSpecifier(node) {
        return node;
    }
    /** @virtual Rewrites a `ViewKeyword` node. */
    rewriteViewKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `VirtualKeyword` node. */
    rewriteVirtualKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `WeeksKeyword` node. */
    rewriteWeeksKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `WeiKeyword` node. */
    rewriteWeiKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `WhileKeyword` node. */
    rewriteWhileKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `Whitespace` node. */
    rewriteWhitespace(node) {
        return node;
    }
    /** @virtual Rewrites a `YearsKeyword` node. */
    rewriteYearsKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `YulAbstractKeyword` node. */
    rewriteYulAbstractKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `YulAfterKeyword` node. */
    rewriteYulAfterKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `YulAliasKeyword` node. */
    rewriteYulAliasKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `YulAnonymousKeyword` node. */
    rewriteYulAnonymousKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `YulApplyKeyword` node. */
    rewriteYulApplyKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `YulAsKeyword` node. */
    rewriteYulAsKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `YulAssemblyKeyword` node. */
    rewriteYulAssemblyKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `YulAutoKeyword` node. */
    rewriteYulAutoKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `YulBoolKeyword` node. */
    rewriteYulBoolKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `YulBreakKeyword` node. */
    rewriteYulBreakKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `YulBytesKeyword` node. */
    rewriteYulBytesKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `YulCallDataKeyword` node. */
    rewriteYulCallDataKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `YulCaseKeyword` node. */
    rewriteYulCaseKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `YulCatchKeyword` node. */
    rewriteYulCatchKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `YulConstantKeyword` node. */
    rewriteYulConstantKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `YulConstructorKeyword` node. */
    rewriteYulConstructorKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `YulContinueKeyword` node. */
    rewriteYulContinueKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `YulContractKeyword` node. */
    rewriteYulContractKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `YulCopyOfKeyword` node. */
    rewriteYulCopyOfKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `YulDaysKeyword` node. */
    rewriteYulDaysKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `YulDecimalLiteral` node. */
    rewriteYulDecimalLiteral(node) {
        return node;
    }
    /** @virtual Rewrites a `YulDefaultKeyword` node. */
    rewriteYulDefaultKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `YulDefineKeyword` node. */
    rewriteYulDefineKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `YulDeleteKeyword` node. */
    rewriteYulDeleteKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `YulDoKeyword` node. */
    rewriteYulDoKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `YulElseKeyword` node. */
    rewriteYulElseKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `YulEmitKeyword` node. */
    rewriteYulEmitKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `YulEnumKeyword` node. */
    rewriteYulEnumKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `YulEtherKeyword` node. */
    rewriteYulEtherKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `YulEventKeyword` node. */
    rewriteYulEventKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `YulExternalKeyword` node. */
    rewriteYulExternalKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `YulFallbackKeyword` node. */
    rewriteYulFallbackKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `YulFalseKeyword` node. */
    rewriteYulFalseKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `YulFinalKeyword` node. */
    rewriteYulFinalKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `YulFinneyKeyword` node. */
    rewriteYulFinneyKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `YulFixedKeyword` node. */
    rewriteYulFixedKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `YulForKeyword` node. */
    rewriteYulForKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `YulFunctionKeyword` node. */
    rewriteYulFunctionKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `YulGweiKeyword` node. */
    rewriteYulGweiKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `YulHexKeyword` node. */
    rewriteYulHexKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `YulHexLiteral` node. */
    rewriteYulHexLiteral(node) {
        return node;
    }
    /** @virtual Rewrites a `YulHoursKeyword` node. */
    rewriteYulHoursKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `YulIdentifier` node. */
    rewriteYulIdentifier(node) {
        return node;
    }
    /** @virtual Rewrites a `YulIfKeyword` node. */
    rewriteYulIfKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `YulImmutableKeyword` node. */
    rewriteYulImmutableKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `YulImplementsKeyword` node. */
    rewriteYulImplementsKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `YulImportKeyword` node. */
    rewriteYulImportKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `YulInKeyword` node. */
    rewriteYulInKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `YulIndexedKeyword` node. */
    rewriteYulIndexedKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `YulInlineKeyword` node. */
    rewriteYulInlineKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `YulIntKeyword` node. */
    rewriteYulIntKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `YulInterfaceKeyword` node. */
    rewriteYulInterfaceKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `YulInternalKeyword` node. */
    rewriteYulInternalKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `YulIsKeyword` node. */
    rewriteYulIsKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `YulLeaveKeyword` node. */
    rewriteYulLeaveKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `YulLetKeyword` node. */
    rewriteYulLetKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `YulLibraryKeyword` node. */
    rewriteYulLibraryKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `YulMacroKeyword` node. */
    rewriteYulMacroKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `YulMappingKeyword` node. */
    rewriteYulMappingKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `YulMatchKeyword` node. */
    rewriteYulMatchKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `YulMemoryKeyword` node. */
    rewriteYulMemoryKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `YulMinutesKeyword` node. */
    rewriteYulMinutesKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `YulModifierKeyword` node. */
    rewriteYulModifierKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `YulMutableKeyword` node. */
    rewriteYulMutableKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `YulNewKeyword` node. */
    rewriteYulNewKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `YulNullKeyword` node. */
    rewriteYulNullKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `YulOfKeyword` node. */
    rewriteYulOfKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `YulOverrideKeyword` node. */
    rewriteYulOverrideKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `YulPartialKeyword` node. */
    rewriteYulPartialKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `YulPayableKeyword` node. */
    rewriteYulPayableKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `YulPragmaKeyword` node. */
    rewriteYulPragmaKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `YulPrivateKeyword` node. */
    rewriteYulPrivateKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `YulPromiseKeyword` node. */
    rewriteYulPromiseKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `YulPublicKeyword` node. */
    rewriteYulPublicKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `YulPureKeyword` node. */
    rewriteYulPureKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `YulReceiveKeyword` node. */
    rewriteYulReceiveKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `YulReferenceKeyword` node. */
    rewriteYulReferenceKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `YulRelocatableKeyword` node. */
    rewriteYulRelocatableKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `YulReturnsKeyword` node. */
    rewriteYulReturnsKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `YulSealedKeyword` node. */
    rewriteYulSealedKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `YulSecondsKeyword` node. */
    rewriteYulSecondsKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `YulSizeOfKeyword` node. */
    rewriteYulSizeOfKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `YulStaticKeyword` node. */
    rewriteYulStaticKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `YulStorageKeyword` node. */
    rewriteYulStorageKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `YulStringKeyword` node. */
    rewriteYulStringKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `YulStructKeyword` node. */
    rewriteYulStructKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `YulSuperKeyword` node. */
    rewriteYulSuperKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `YulSupportsKeyword` node. */
    rewriteYulSupportsKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `YulSwitchKeyword` node. */
    rewriteYulSwitchKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `YulSzaboKeyword` node. */
    rewriteYulSzaboKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `YulThisKeyword` node. */
    rewriteYulThisKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `YulThrowKeyword` node. */
    rewriteYulThrowKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `YulTrueKeyword` node. */
    rewriteYulTrueKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `YulTryKeyword` node. */
    rewriteYulTryKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `YulTypeDefKeyword` node. */
    rewriteYulTypeDefKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `YulTypeKeyword` node. */
    rewriteYulTypeKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `YulTypeOfKeyword` node. */
    rewriteYulTypeOfKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `YulUfixedKeyword` node. */
    rewriteYulUfixedKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `YulUintKeyword` node. */
    rewriteYulUintKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `YulUncheckedKeyword` node. */
    rewriteYulUncheckedKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `YulUsingKeyword` node. */
    rewriteYulUsingKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `YulVarKeyword` node. */
    rewriteYulVarKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `YulViewKeyword` node. */
    rewriteYulViewKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `YulVirtualKeyword` node. */
    rewriteYulVirtualKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `YulWeeksKeyword` node. */
    rewriteYulWeeksKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `YulWeiKeyword` node. */
    rewriteYulWeiKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `YulWhileKeyword` node. */
    rewriteYulWhileKeyword(node) {
        return node;
    }
    /** @virtual Rewrites a `YulYearsKeyword` node. */
    rewriteYulYearsKeyword(node) {
        return node;
    }
    /** @virtual Rewrites an `Unrecognized` node. */
    rewriteUnrecognized(node) {
        return node;
    }
    /** @virtual Rewrites a `Missing` node. */
    rewriteMissing(node) {
        return node;
    }
    /** Rewrites all the children of a given non-terminal node. */
    rewriteChildren(node) {
        let newChildren = undefined;
        const children = node.children();
        children.forEach((child, index) => {
            const newChild = this.rewriteNode(child.node);
            if (newChild == undefined) {
                // node was removed. if `newChildren` is set, just skip this one
                // otherwise, copy the first ones from `children` (but the last)
                if (newChildren == undefined) {
                    newChildren = children.slice(0, index);
                }
            }
            else if (newChild.id != child.node.id) {
                // node has changed, produce new edge
                let edge;
                switch (newChild.type) {
                    case NodeType.TerminalNode:
                        edge = Edge.createWithTerminal(child.label, newChild);
                        break;
                    case NodeType.NonterminalNode:
                        edge = Edge.createWithNonterminal(child.label, newChild);
                        break;
                }
                if (newChildren == undefined) {
                    newChildren = children.slice(0, index);
                }
                newChildren.push(edge);
            }
            else if (newChildren != undefined) {
                newChildren.push(child);
            }
        });
        if (newChildren != undefined) {
            return NonterminalNode.create(node.kind, newChildren);
        }
        else {
            return node;
        }
    }
}
//# sourceMappingURL=rewriter.generated.mjs.map