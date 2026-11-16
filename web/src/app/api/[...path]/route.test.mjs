import assert from "node:assert/strict";
import { readFile } from "node:fs/promises";
import { fileURLToPath } from "node:url";
import test from "node:test";

test("API proxy preserves content length for partial video responses", async () => {
    const routePath = fileURLToPath(new URL("./route.ts", import.meta.url));
    const source = await readFile(routePath, "utf8");

    assert.match(source, /const isMediaResponse = response\.status === 206 \|\| response\.headers\.get\("content-type"\)\?\.startsWith\("video\/"\)/);
    assert.match(source, /if \(!isMediaResponse\) headers\.delete\("content-length"\)/);
});
