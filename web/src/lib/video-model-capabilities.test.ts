import assert from "node:assert/strict";
import test from "node:test";

import { museVideoReferenceLimitError } from "./video-model-capabilities";

test("Muse Video rejects more than ten reference assets before submission", () => {
    assert.equal(museVideoReferenceLimitError("muse-video", 11), "Muse Video 最多支持 10 个参考素材，当前已选择 11 个。请移除 1 个后重试。");
});

test("Muse Video accepts ten or fewer reference images", () => {
    assert.equal(museVideoReferenceLimitError("muse-video", 10), "");
    assert.equal(museVideoReferenceLimitError("other-video", 11), "");
});
