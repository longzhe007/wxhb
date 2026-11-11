import assert from "node:assert/strict";
import test from "node:test";

import { museVideoReferenceLimitError } from "./video-model-capabilities";

test("Muse Video rejects more than ten reference images before submission", () => {
    assert.equal(museVideoReferenceLimitError("muse-video", 11), "Muse Video 最多支持 10 张参考图，当前已选择 11 张。请移除 1 张或将多张角色设定合成为一张参考图。");
});

test("Muse Video accepts ten or fewer reference images", () => {
    assert.equal(museVideoReferenceLimitError("muse-video", 10), "");
    assert.equal(museVideoReferenceLimitError("other-video", 11), "");
});
