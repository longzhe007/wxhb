import { cp, stat } from "node:fs/promises";
import { dirname, join, resolve } from "node:path";
import { fileURLToPath } from "node:url";

export async function copyStandaloneAssets(webDir) {
    const nextDir = join(webDir, ".next");
    const standaloneDir = join(nextDir, "standalone");

    await cp(join(nextDir, "static"), join(standaloneDir, ".next", "static"), { recursive: true, force: true });

    try {
        await stat(join(webDir, "public"));
        await cp(join(webDir, "public"), join(standaloneDir, "public"), { recursive: true, force: true });
    } catch (error) {
        if (error && typeof error === "object" && "code" in error && error.code === "ENOENT") return;
        throw error;
    }
}

const scriptPath = fileURLToPath(import.meta.url);
if (process.argv[1] && resolve(process.argv[1]) === scriptPath) {
    const webDir = dirname(dirname(scriptPath));
    await copyStandaloneAssets(webDir);
}
