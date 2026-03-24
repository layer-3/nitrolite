import { Cursor } from "../cst/index.mjs";
import { UserFileLocation, BuiltInLocation } from "./index.mjs";
/**
 * Asserts that `location` is a `UserFileLocation`.
 *
 * If a `fileId` value is provided, it will also assert that it matches its file ID.
 *
 * If a `cursor` value is provided, it will also assert that it points to the same node the cursor is pointing at.
 */
export declare function assertUserFileLocation(location: unknown, fileId?: string, cursor?: Cursor): asserts location is UserFileLocation;
/**
 * Asserts that `location` is a `BuiltInLocation`.
 */
export declare function assertBuiltInLocation(location: unknown): asserts location is BuiltInLocation;
//# sourceMappingURL=assertions.d.mts.map