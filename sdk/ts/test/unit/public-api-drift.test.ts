import * as publicApi from '../../src/index.js';
import path from 'node:path';
import { fileURLToPath } from 'node:url';
import ts from 'typescript';

const testDir = path.dirname(fileURLToPath(import.meta.url));
const packageRoot = path.resolve(testDir, '../..');

const FORMAT_FLAGS =
    ts.TypeFormatFlags.NoTruncation |
    ts.TypeFormatFlags.UseSingleQuotesForStringLiteralType |
    ts.TypeFormatFlags.WriteArrayAsGenericType |
    ts.TypeFormatFlags.UseAliasDefinedOutsideCurrentScope;

type PublicApiMember = {
    name: string;
    kind: string;
    signatures?: string[];
    constructors?: string[];
    properties?: string[];
    staticProperties?: string[];
    members?: string[];
    type?: string;
    declaration?: string;
};

function normalizeText(text: string): string {
    return text.replace(/\s+/g, ' ').trim();
}

function createPackageProgram() {
    const configPath = ts.findConfigFile(packageRoot, ts.sys.fileExists, 'tsconfig.json');
    if (!configPath) throw new Error(`tsconfig.json not found under ${packageRoot}`);

    const configFile = ts.readConfigFile(configPath, ts.sys.readFile);
    if (configFile.error) {
        throw new Error(ts.flattenDiagnosticMessageText(configFile.error.messageText, '\n'));
    }

    const parsed = ts.parseJsonConfigFileContent(configFile.config, ts.sys, packageRoot);
    return ts.createProgram(parsed.fileNames, parsed.options);
}

function declarationKind(declaration: ts.Declaration): string {
    if (ts.isClassDeclaration(declaration)) return 'class';
    if (ts.isInterfaceDeclaration(declaration)) return 'interface';
    if (ts.isFunctionDeclaration(declaration)) return 'function';
    if (ts.isEnumDeclaration(declaration)) return 'enum';
    if (ts.isTypeAliasDeclaration(declaration)) return 'type';
    if (ts.isVariableDeclaration(declaration)) return 'const';
    return ts.SyntaxKind[declaration.kind] ?? 'unknown';
}

function signaturesForType(
    checker: ts.TypeChecker,
    type: ts.Type,
    declaration: ts.Declaration
): string[] {
    return type
        .getCallSignatures()
        .map((signature) => checker.signatureToString(signature, declaration, FORMAT_FLAGS))
        .sort();
}

function isPrivateOrProtected(declaration: ts.Declaration): boolean {
    const flags = ts.getCombinedModifierFlags(declaration as ts.Declaration);
    return Boolean(flags & (ts.ModifierFlags.Private | ts.ModifierFlags.Protected));
}

function propertiesForType(
    checker: ts.TypeChecker,
    type: ts.Type,
    declaration: ts.Declaration
): string[] {
    return checker
        .getPropertiesOfType(type)
        .flatMap((property) => {
            const propertyDeclaration = property.valueDeclaration ?? property.declarations?.[0] ?? declaration;
            if (isPrivateOrProtected(propertyDeclaration)) return [];

            const propertyType = checker.getTypeOfSymbolAtLocation(property, propertyDeclaration);
            const signatures = signaturesForType(checker, propertyType, propertyDeclaration);
            if (signatures.length > 0) {
                return [`${property.getName()}: ${signatures.join(' | ')}`];
            }
            if (
                (ts.isPropertySignature(propertyDeclaration) ||
                    ts.isPropertyDeclaration(propertyDeclaration)) &&
                propertyDeclaration.type
            ) {
                return [`${property.getName()}: ${normalizeText(propertyDeclaration.type.getText())}`];
            }
            return [
                `${property.getName()}: ${checker.typeToString(propertyType, propertyDeclaration, FORMAT_FLAGS)}`,
            ];
        })
        .sort();
}

function enumMembers(declaration: ts.EnumDeclaration): string[] {
    return declaration.members.map((member) => {
        const initializer = member.initializer ? normalizeText(member.initializer.getText()) : '<auto>';
        return `${member.name.getText()} = ${initializer}`;
    });
}

function serializePublicApi(): PublicApiMember[] {
    const program = createPackageProgram();
    const checker = program.getTypeChecker();
    const entrypoint = program.getSourceFile(path.join(packageRoot, 'src/index.ts'));
    if (!entrypoint) throw new Error('src/index.ts not found in program');

    const moduleSymbol = checker.getSymbolAtLocation(entrypoint);
    if (!moduleSymbol) throw new Error('src/index.ts module symbol not found');

    return checker
        .getExportsOfModule(moduleSymbol)
        .filter((symbol) => symbol.getName() !== '__esModule')
        .map((exportedSymbol) => {
            const symbol =
                exportedSymbol.flags & ts.SymbolFlags.Alias
                    ? checker.getAliasedSymbol(exportedSymbol)
                    : exportedSymbol;
            const declaration = symbol.getDeclarations()?.[0];
            if (!declaration) {
                return {
                    name: exportedSymbol.getName(),
                    kind: 'unknown',
                };
            }

            const kind = declarationKind(declaration);
            const member: PublicApiMember = {
                name: exportedSymbol.getName(),
                kind,
            };

            if (ts.isClassDeclaration(declaration)) {
                const staticType = checker.getTypeOfSymbolAtLocation(symbol, declaration);
                const instanceType = checker.getDeclaredTypeOfSymbol(symbol);
                member.constructors = staticType
                    .getConstructSignatures()
                    .map((signature) => checker.signatureToString(signature, declaration, FORMAT_FLAGS))
                    .sort();
                member.properties = propertiesForType(checker, instanceType, declaration);
                member.staticProperties = propertiesForType(checker, staticType, declaration).filter(
                    (property) => !['length', 'name', 'prototype'].some((skip) => property.startsWith(`${skip}:`))
                );
            } else if (ts.isInterfaceDeclaration(declaration)) {
                const type = checker.getDeclaredTypeOfSymbol(symbol);
                member.properties = propertiesForType(checker, type, declaration);
                member.signatures = signaturesForType(checker, type, declaration);
            } else if (ts.isFunctionDeclaration(declaration)) {
                member.signatures = signaturesForType(
                    checker,
                    checker.getTypeOfSymbolAtLocation(symbol, declaration),
                    declaration
                );
            } else if (ts.isEnumDeclaration(declaration)) {
                member.members = enumMembers(declaration);
            } else if (ts.isTypeAliasDeclaration(declaration)) {
                member.declaration = normalizeText(declaration.type.getText());
                member.type = checker.typeToString(
                    checker.getTypeFromTypeNode(declaration.type),
                    declaration,
                    FORMAT_FLAGS
                );
            } else if (ts.isVariableDeclaration(declaration)) {
                member.type = checker.typeToString(
                    checker.getTypeOfSymbolAtLocation(symbol, declaration),
                    declaration,
                    FORMAT_FLAGS
                );
            }

            return member;
        })
        .sort((a, b) => a.name.localeCompare(b.name));
}

describe('SDK public runtime API drift guard', () => {
    it('keeps root runtime exports intentional', () => {
        expect(Object.keys(publicApi).sort()).toMatchSnapshot();
    });

    it('keeps root TypeScript public API signatures intentional', () => {
        expect(serializePublicApi()).toMatchSnapshot();
    });

    it('proves adversarial public export removal is observable', () => {
        const exports = new Set(Object.keys(publicApi));
        exports.delete('Client');

        expect(exports.has('Client')).toBe(false);
    });

    it('proves adversarial public signature changes are observable', () => {
        const api = serializePublicApi();
        const client = api.find((member) => member.name === 'Client');
        expect(client?.properties?.some((property) => property.includes('ping:'))).toBe(true);

        const mutated = api.map((member) =>
            member.name === 'Client'
                ? {
                    ...member,
                    properties: member.properties?.filter((property) => !property.includes('ping:')),
                }
                : member
        );
        const mutatedClient = mutated.find((member) => member.name === 'Client');

        expect(mutatedClient?.properties?.some((property) => property.includes('ping:'))).toBe(false);
    });

    it('proves adversarial type-only export removal is observable', () => {
        const api = serializePublicApi();
        expect(api.some((member) => member.name === 'Config' && member.kind === 'interface')).toBe(true);

        const mutated = api.filter((member) => member.name !== 'Config');
        expect(mutated.some((member) => member.name === 'Config')).toBe(false);
    });

    it('proves adversarial function parameter changes are observable', () => {
        const api = serializePublicApi();
        const packer = api.find((member) => member.name === 'packAppStateUpdateV1');
        const original = packer?.signatures?.[0] ?? '';
        expect(original).toContain('stateUpdate: AppStateUpdateV1');

        const mutated = original.replace('stateUpdate: AppStateUpdateV1', 'stateUpdate: unknown');
        expect(mutated).not.toEqual(original);
    });

    it('proves adversarial enum value changes are observable', () => {
        const api = serializePublicApi();
        const intent = api.find((member) => member.name === 'AppStateUpdateIntent');
        const original = intent?.members?.join('|') ?? '';
        expect(original).toContain('Deposit');

        const mutated = original.replace('Deposit', 'DepositChanged');
        expect(mutated).not.toEqual(original);
    });

    it('proves adversarial public export additions are observable', () => {
        const api = serializePublicApi();
        expect(api.some((member) => member.name === '__FakeExport')).toBe(false);

        const mutated = [...api, { name: '__FakeExport', kind: 'function' }];
        expect(mutated.some((member) => member.name === '__FakeExport')).toBe(true);
    });
});
