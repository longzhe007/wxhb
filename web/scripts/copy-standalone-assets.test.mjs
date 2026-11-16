import assert from "node:assert/strict";
import { mkdtemp, mkdir, readFile, rm, writeFile } from "node:fs/promises";
import { tmpdir } from "node:os";
import { join } from "node:path";
import test from "node:test";

import { copyStandaloneAssets } from "./copy-standalone-assets.mjs";

test("copies Next static and public assets into the standalone server", async () => {
    const webDir = await mkdtemp(join(tmpdir(), "infinite-canvas-web-"));
    try {
        await mkdir(join(webDir, ".next", "static", "chunks"), { recursive: true });
        await mkdir(join(webDir, ".next", "standalone"), { recursive: true });
        await mkdir(join(webDir, "public"), { recursive: true });
        await writeFile(join(webDir, ".next", "static", "chunks", "app.css"), "body { color: red; }");
        await writeFile(join(webDir, "public", "logo.svg"), "<svg />");

        await copyStandaloneAssets(webDir);

        assert.equal(await readFile(join(webDir, ".next", "standalone", ".next", "static", "chunks", "app.css"), "utf8"), "body { color: red; }");
        assert.equal(await readFile(join(webDir, ".next", "standalone", "public", "logo.svg"), "utf8"), "<svg />");
    } finally {
        await rm(webDir, { recursive: true, force: true });
    }
});
