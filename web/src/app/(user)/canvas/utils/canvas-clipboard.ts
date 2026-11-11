import { CanvasNodeType, type CanvasConnection, type CanvasNodeData } from "../types";

export type CanvasClipboard = {
    nodes: CanvasNodeData[];
    connections: CanvasConnection[];
};

export function buildCanvasClipboard(nodes: CanvasNodeData[], connections: CanvasConnection[], selectedIds: ReadonlySet<string>): CanvasClipboard | null {
    if (!selectedIds.size) return null;

    const copiedIds = new Set(selectedIds);
    nodes.forEach((node) => {
        if (node.type !== CanvasNodeType.Group || !selectedIds.has(node.id)) return;
        nodes.forEach((child) => {
            if (child.metadata?.groupId === node.id) copiedIds.add(child.id);
        });
    });

    const copiedNodes = nodes
        .filter((node) => copiedIds.has(node.id))
        .map((node) => ({
            ...node,
            position: { ...node.position },
            metadata: node.metadata ? { ...node.metadata } : undefined,
        }));
    if (!copiedNodes.length) return null;

    return {
        nodes: copiedNodes,
        // Keep copied nodes connected to their existing input references.
        connections: connections.filter((connection) => copiedIds.has(connection.toNodeId)).map((connection) => ({ ...connection })),
    };
}

export function remapCanvasClipboardConnections(connections: CanvasConnection[], idMap: ReadonlyMap<string, string>, availableNodeIds: ReadonlySet<string>) {
    return connections.flatMap((connection) => {
        const fromNodeId = idMap.get(connection.fromNodeId) || (availableNodeIds.has(connection.fromNodeId) ? connection.fromNodeId : "");
        const toNodeId = idMap.get(connection.toNodeId);
        return fromNodeId && toNodeId ? [{ ...connection, fromNodeId, toNodeId }] : [];
    });
}
