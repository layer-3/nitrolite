import { NonterminalKind, NonterminalNode, TerminalKind, TerminalNode } from "./index.mjs";
/**
 * Asserts that `node` is a `NonterminalNode`.
 *
 * If a `kind` value is provided, it will also assert that it matches its `NonterminalKind`.
 *
 * If a `text` value is provided, it will also be asserted against the node contents.
 */
export declare function assertNonterminalNode(node: unknown, kind?: NonterminalKind, text?: string): asserts node is NonterminalNode;
/**
 * Asserts that `node` is a `TerminalNode`.
 *
 * If a `kind` value is provided, it will also assert that it matches its `TerminalKind`.
 *
 * If a `text` value is provided, it will also be asserted against the node contents.
 */
export declare function assertTerminalNode(node: unknown, kind?: TerminalKind, text?: string): asserts node is TerminalNode;
//# sourceMappingURL=assertions.d.mts.map