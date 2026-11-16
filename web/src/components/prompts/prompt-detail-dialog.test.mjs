import assert from "node:assert/strict";
import { readFile } from "node:fs/promises";
import { fileURLToPath } from "node:url";
import test from "node:test";

test("prompt video player uses the same-origin media proxy", async () => {
    const componentPath = fileURLToPath(new URL("./prompt-detail-dialog.tsx", import.meta.url));
    const source = await readFile(componentPath, "utf8");

    assert.match(source, /\/api\/proxy-media\?url=/);
    assert.doesNotMatch(source, /<video src=\{prompt\.videoUrl\}/);
});
