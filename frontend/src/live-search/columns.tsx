import type { ColumnDef } from "@tanstack/react-table";
import { Button } from "@/components/ui/button";
import { Trash2 } from "lucide-react";
import { Input } from "@/components/ui/input";
import { Checkbox } from "@/components/ui/checkbox";
import React from "react";
import { livesearch } from "../../wailsjs/go/models";

import TradeLink = livesearch.TradeLink;

type ActionsProps = {
  editIdx: number | null;
  editUrl: string;
  editDescription: string;
  setEditUrl: (url: string) => void;
  setEditDescription: (desc: string) => void;
  handleEdit: (idx: number) => void;
  handleSaveEdit: (idx: number) => void;
  handleCancelEdit: () => void;
  handleDelete: (idx: number) => void;
};

export const getColumns = (actions: ActionsProps): ColumnDef<TradeLink>[] => [
  {
    id: "select",
    header: ({ table }) => (
      <Checkbox
        checked={
          table.getIsAllPageRowsSelected() || (table.getIsSomePageRowsSelected() && "indeterminate")
        }
        onCheckedChange={(value) => table.toggleAllPageRowsSelected(!!value)}
        aria-label="Select all"
      />
    ),
    cell: ({ row }) => (
      <Checkbox
        checked={row.getIsSelected()}
        onCheckedChange={(value) => row.toggleSelected(!!value)}
        aria-label="Select row"
      />
    ),
    enableSorting: false,
    enableHiding: false,
  },
  {
    accessorKey: "searchId",
    header: "Search ID",
  },
  {
    accessorKey: "league",
    header: "League",
    cell: ({ row, getValue }) => (
      <span
        className="block max-w-[100px] overflow-hidden whitespace-nowrap text-ellipsis"
        title={getValue() as string}
      >
        {getValue() as string}
      </span>
    ),
  },
  {
    accessorKey: "url",
    header: "URL",
    cell: ({ row, getValue }) =>
      actions.editIdx === row.index ? (
        <Input value={actions.editUrl} onChange={(e) => actions.setEditUrl(e.target.value)} />
      ) : (
        <span
          className="block max-w-[120px] overflow-hidden whitespace-nowrap text-ellipsis"
          title={getValue() as string}
        >
          {getValue() as string}
        </span>
      ),
  },
  {
    accessorKey: "description",
    header: "Description",
    cell: ({ row, getValue }) =>
      actions.editIdx === row.index ? (
        <Input
          value={actions.editDescription}
          onChange={(e) => actions.setEditDescription(e.target.value)}
        />
      ) : (
        <span
          className="block max-w-[100px] overflow-hidden whitespace-nowrap text-ellipsis"
          title={getValue() as string}
        >
          {getValue() as string}
        </span>
      ),
  },
  {
    accessorKey: "status",
    header: "Status",
  },
  {
    id: "actions",
    header: "Actions",
    cell: ({ row }) =>
      actions.editIdx === row.index ? (
        <>
          <Button size="sm" onClick={() => actions.handleSaveEdit(row.index)}>
            Save
          </Button>
          <Button className="ml-2" size="sm" variant="secondary" onClick={actions.handleCancelEdit}>
            Cancel
          </Button>
        </>
      ) : (
        <>
          <Button size="sm" variant="secondary" onClick={() => actions.handleEdit(row.index)}>
            Edit
          </Button>
          <Button
            size="sm"
            variant="destructive"
            className="ml-2"
            onClick={() => actions.handleDelete(row.index)}
            aria-label="Delete"
          >
            <Trash2 className="w-4 h-4" />
          </Button>
        </>
      ),
  },
];
