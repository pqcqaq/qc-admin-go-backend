-- Migration: Remove redundant fields from workflow tables
-- Date: 2025-11-01
-- Description: Remove node_key, next_node_id, parent_node_id from workflow_nodes and edge_key from workflow_edges
--              These fields are no longer needed as:
--              - Frontend now uses database IDs directly (no need for node_key/edge_key)
--              - Node connections are managed through workflow_edges table (no need for next_node_id)
--              - Parallel child relationships are managed through workflow_edges table (no need for parent_node_id)

-- ============================================================
-- Step 1: Remove indexes first (to avoid foreign key issues)
-- ============================================================

-- Remove index on next_node_id from workflow_nodes
DROP INDEX IF EXISTS workflownode_next_node_id ON workflow_nodes;

-- Remove index on parent_node_id from workflow_nodes
DROP INDEX IF EXISTS workflownode_parent_node_id ON workflow_nodes;

-- Remove unique index on (application_id, node_key) from workflow_nodes
DROP INDEX IF EXISTS workflownode_application_id_node_key ON workflow_nodes;

-- Remove unique index on (application_id, edge_key) from workflow_edges
DROP INDEX IF EXISTS workflowedge_application_id_edge_key ON workflow_edges;

-- ============================================================
-- Step 2: Remove columns
-- ============================================================

-- Remove node_key from workflow_nodes
ALTER TABLE workflow_nodes DROP COLUMN IF EXISTS node_key;

-- Remove next_node_id from workflow_nodes
ALTER TABLE workflow_nodes DROP COLUMN IF EXISTS next_node_id;

-- Remove parent_node_id from workflow_nodes
ALTER TABLE workflow_nodes DROP COLUMN IF EXISTS parent_node_id;

-- Remove edge_key from workflow_edges
ALTER TABLE workflow_edges DROP COLUMN IF EXISTS edge_key;

-- ============================================================
-- Migration Complete
-- ============================================================
-- The following fields have been removed:
-- - workflow_nodes.node_key (was used for frontend temporary ID mapping)
-- - workflow_nodes.next_node_id (was used for old connection system)
-- - workflow_nodes.parent_node_id (was used for parallel child relationships)
-- - workflow_edges.edge_key (was used for frontend temporary ID mapping)
--
-- All node connections are now managed through the workflow_edges table
-- using source_node_id and target_node_id fields.
--
-- Parallel child relationships are also managed through workflow_edges
-- with type="parallel" and appropriate sourceHandle values.
--
-- The frontend now uses database IDs directly instead of temporary IDs.
-- ============================================================

