import assert from "node:assert/strict";
import { readFile } from "node:fs/promises";
import { fileURLToPath } from "node:url";
import test from "node:test";

test("homepage prompt showcase uses the empty-cover fallback", async () => {
    const pagePath = fileURLToPath(new URL("./page.tsx", import.meta.url));
    const source = await readFile(pagePath, "utf8");

    assert.match(source, /import \{ PromptCover \} from "@\/components\/prompts\/prompt-cover"/);
    assert.match(source, /<PromptCover coverUrl=\{item\.coverUrl\} title=\{item\.title\}/);
    assert.doesNotMatch(source, /<img src=\{item\.coverUrl\}/);
});
