import { Button } from "@/components/ui/button";
import { Checkbox } from "@/components/ui/checkbox";
import { Input } from "@/components/ui/input";
import type { ColumnDef } from "@tanstack/react-table";
import { Trash2 } from "lucide-react";
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
  handleSelect(idx: number, selected: boolean): void;
  data: TradeLink[];
};

export const getColumns = (actions: ActionsProps): ColumnDef<TradeLink>[] => [
  {
    id: "select",
    header: () => {
      // Determine if all or some are selected
      const allSelected = actions.data.every((l) => l.selected);
      const someSelected = actions.data.some((l) => l.selected) && !allSelected;
      return (
        <Checkbox
          checked={allSelected ? true : someSelected ? "indeterminate" : false}
          onCheckedChange={(value) => {
            actions.data.forEach((_, idx) => {
              actions.handleSelect(idx, !!value);
            });
          }}
          aria-label="Select all"
        />
      );
    },
    cell: ({ row }) => (
      <Checkbox
        checked={row.original.selected}
        onCheckedChange={(checked) => {
          console.log("Row index:", row.index, "Checked:", checked);
          actions.handleSelect(row.index, !!checked);
        }}
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
    cell: ({ getValue }) => (
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
        <Input
          className="min-w-[200px]"
          value={actions.editUrl}
          onChange={(e) => actions.setEditUrl(e.target.value)}
        />
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
          className="min-w-[150px]"
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
        <div className="flex items-center gap-2">
          <Button size="sm" variant="secondary" onClick={() => actions.handleEdit(row.index)}>
            Edit
          </Button>
          <Button
            size="sm"
            variant="destructive"
            onClick={() => actions.handleDelete(row.index)}
            aria-label="Delete"
          >
            <Trash2 className="w-4 h-4" />
          </Button>
        </div>
      ),
  },
];
