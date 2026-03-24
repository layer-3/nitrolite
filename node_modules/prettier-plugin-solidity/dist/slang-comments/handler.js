import handlers from './handlers/index.js';
function ownLine(comment, text, options, ast, isLastComment) {
    const { precedingNode, enclosingNode, followingNode } = comment;
    const handlerArguments = {
        text,
        precedingNode,
        enclosingNode,
        followingNode,
        comment,
        ast,
        isLastComment
    };
    return handlers.some((handler) => handler(handlerArguments));
}
function endOfLine(comment, text, options, ast, isLastComment) {
    const { precedingNode, enclosingNode, followingNode } = comment;
    const handlerArguments = {
        text,
        precedingNode,
        enclosingNode,
        followingNode,
        comment,
        ast,
        isLastComment
    };
    return handlers.some((handler) => handler(handlerArguments));
}
function remaining(comment, text, options, ast, isLastComment) {
    const { precedingNode, enclosingNode, followingNode } = comment;
    const handlerArguments = {
        text,
        precedingNode,
        enclosingNode,
        followingNode,
        comment,
        ast,
        isLastComment
    };
    return handlers.some((handler) => handler(handlerArguments));
}
export const handleComments = { ownLine, endOfLine, remaining };
//# sourceMappingURL=handler.js.map