import assert from "node:assert/strict";
import test from "node:test";

import { buildCanvasClipboard, duplicateCanvasNode, remapCanvasClipboardConnections } from "./canvas-clipboard";
import { CanvasNodeType, type CanvasConnection, type CanvasNodeData } from "../types";

function node(id: string, type = CanvasNodeType.Image): CanvasNodeData {
    return { id, type, title: id, position: { x: 0, y: 0 }, width: 100, height: 100 };
}

function connection(id: string, fromNodeId: string, toNodeId: string): CanvasConnection {
    return { id, fromNodeId, toNodeId };
}

test("copying a generated node preserves its unselected reference input", () => {
    const clipboard = buildCanvasClipboard(
        [node("reference"), node("generated"), node("downstream")],
        [connection("input", "reference", "generated"), connection("output", "generated", "downstream")],
        new Set(["generated"]),
    );

    assert.deepEqual(clipboard?.connections, [connection("input", "reference", "generated")]);

    const remapped = remapCanvasClipboardConnections(clipboard!.connections, new Map([["generated", "generated-copy"]]), new Set(["reference", "generated-copy", "downstream"]));
    assert.deepEqual(remapped, [connection("input", "reference", "generated-copy")]);
});

test("duplicating from the node menu preserves its input references", () => {
    const result = duplicateCanvasNode(
        [node("reference"), node("generated"), node("downstream")],
        [connection("input", "reference", "generated"), connection("output", "generated", "downstream")],
        "generated",
        "generated-duplicate",
    );

    assert.equal(result?.node.id, "generated-duplicate");
    assert.deepEqual(result?.connections, [connection("input", "reference", "generated-duplicate")]);
});

test("copying multiple nodes preserves remapped internal connections", () => {
    const clipboard = buildCanvasClipboard(
        [node("reference"), node("first"), node("second"), node("downstream")],
        [connection("input", "reference", "first"), connection("internal", "first", "second"), connection("output", "second", "downstream")],
        new Set(["first", "second"]),
    );

    assert.deepEqual(clipboard?.connections, [connection("input", "reference", "first"), connection("internal", "first", "second")]);

    const remapped = remapCanvasClipboardConnections(
        clipboard!.connections,
        new Map([["first", "first-copy"], ["second", "second-copy"]]),
        new Set(["reference", "first-copy", "second-copy", "downstream"]),
    );
    assert.deepEqual(remapped, [connection("input", "reference", "first-copy"), connection("internal", "first-copy", "second-copy")]);
});
