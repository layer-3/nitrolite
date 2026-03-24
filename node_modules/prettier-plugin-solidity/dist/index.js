import * as comments from './comments/index.js';
import { handleComments, printComment } from './slang-comments/index.js';
import massageAstNode from './clean.js';
import loc from './loc.js';
import options from './options.js';
import antlrParse from './parser.js';
import antlrPrint from './printer.js';
import slangParse from './slangSolidityParser.js';
import slangPrint from './slangPrinter.js';
import { isBlockComment, isComment } from './slang-utils/is-comment.js';
import { locEnd, locStart } from './slang-utils/loc.js';
const slangParserId = 'slang';
const antlrParserId = 'antlr';
const slangAstId = 'slang-ast';
const antlrAstId = 'antlr-ast';
// https://prettier.io/docs/en/plugins.html#languages
// https://github.com/github-linguist/linguist/blob/master/lib/linguist/languages.yml
const languages = [
    {
        linguistLanguageId: 237469032,
        name: 'Solidity',
        aceMode: 'text',
        tmScope: 'source.solidity',
        extensions: ['.sol'],
        parsers: [slangParserId, antlrParserId],
        vscodeLanguageIds: ['solidity']
    }
];
// https://prettier.io/docs/en/plugins.html#parsers
const antlrParser = Object.assign({ astFormat: antlrAstId, parse: antlrParse }, loc);
const slangParser = {
    astFormat: slangAstId,
    parse: slangParse,
    locStart,
    locEnd
};
const parsers = {
    [slangParserId]: slangParser,
    [antlrParserId]: antlrParser
};
const antlrCanAttachComment = ({ type }) => typeof type === 'string' && type !== 'BlockComment' && type !== 'LineComment';
const canAttachComment = (node) => typeof node !== 'string' &&
    node !== undefined &&
    node.kind && // Make sure it's not Location
    !isComment(node);
// https://prettier.io/docs/en/plugins.html#printers
const antlrPrinter = {
    canAttachComment: antlrCanAttachComment,
    handleComments: {
        ownLine: comments.solidityHandleOwnLineComment,
        endOfLine: comments.solidityHandleEndOfLineComment,
        remaining: comments.solidityHandleRemainingComment
    },
    isBlockComment: comments.isBlockComment,
    massageAstNode,
    print: antlrPrint,
    printComment: comments.printComment
};
const slangPrinter = {
    canAttachComment,
    handleComments,
    isBlockComment,
    massageAstNode,
    print: slangPrint,
    printComment
};
const printers = {
    [slangAstId]: slangPrinter,
    [antlrAstId]: antlrPrinter
};
// https://prettier.io/docs/en/plugins.html#defaultoptions
const defaultOptions = {
    bracketSpacing: false,
    tabWidth: 4
};
export default { languages, parsers, printers, options, defaultOptions };
//# sourceMappingURL=index.js.map