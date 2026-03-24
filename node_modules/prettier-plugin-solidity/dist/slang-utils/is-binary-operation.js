import { NonterminalKind } from '@nomicfoundation/slang/cst';
import { createKindCheckFunction } from './create-kind-check-function.js';
export const isBinaryOperation = createKindCheckFunction([
    NonterminalKind.AdditiveExpression,
    NonterminalKind.MultiplicativeExpression,
    NonterminalKind.ExponentiationExpression,
    NonterminalKind.AssignmentExpression,
    NonterminalKind.BitwiseAndExpression,
    NonterminalKind.BitwiseOrExpression,
    NonterminalKind.BitwiseXorExpression,
    NonterminalKind.EqualityExpression,
    NonterminalKind.InequalityExpression,
    NonterminalKind.AndExpression,
    NonterminalKind.OrExpression,
    NonterminalKind.ShiftExpression
]);
//# sourceMappingURL=is-binary-operation.js.map