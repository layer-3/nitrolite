import { Edge } from "./index.mjs";
/** Create a new `Edge` with the label `label` and node `node`. */
export function createEdge(label, node) {
    if (node.isNonterminalNode()) {
        return Edge.createWithNonterminal(label, node.asNonterminalNode());
    }
    else {
        return Edge.createWithTerminal(label, node.asTerminalNode());
    }
}
//# sourceMappingURL=extensions.mjs.map