import React, { useEffect, useState } from "react";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { Card, CardHeader, CardTitle, CardContent, CardFooter } from "@/components/ui/card";
import { Checkbox } from "@/components/ui/checkbox";
import { Trash2 } from "lucide-react";
import {
  Table,
  TableHeader,
  TableRow,
  TableHead,
  TableBody,
  TableCell,
} from "@/components/ui/table";
import { toast } from "sonner";
import { listTradeLinks, addTradeLink, updateTradeLinks } from "@/services/liveSearchService";
import { livesearch } from "../../wailsjs/go/models";

import TradeLink = livesearch.TradeLink;

export default function LiveSearch() {
  const [url, setUrl] = useState("");
  const [description, setDescription] = useState("");
  const [links, setLinks] = useState<TradeLink[]>([]);
  const [editIdx, setEditIdx] = useState<number | null>(null);
  const [editUrl, setEditUrl] = useState("");
  const [editDescription, setEditDescription] = useState("");

  useEffect(() => {
    listTradeLinks().then(setLinks);
  }, []);

  const handleAdd = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!url) return;
    await addTradeLink(url, description);
    setUrl("");
    setDescription("");
    setLinks(await listTradeLinks());
    toast("Link added!");
  };

  const handleSelect = (idx: number, checked: boolean) => {
    const updated = links.map((l, i) => (i === idx ? { ...l, selected: checked } : l));
    setLinks(updated);
    updateTradeLinks(updated);
  };

  const handleDelete = async (idx: number) => {
    const updated = links.filter((_, i) => i !== idx);
    setLinks(updated);
    await updateTradeLinks(updated);
    toast("Link deleted!");
  };

  const handleEdit = (idx: number) => {
    setEditIdx(idx);
    setEditUrl(links[idx].url);
    setEditDescription(links[idx].description || "");
  };

  const handleSaveEdit = async (idx: number) => {
    const updated = links.map((l, i) =>
      i === idx ? { ...l, url: editUrl, description: editDescription } : l,
    );
    setLinks(updated);
    await updateTradeLinks(updated);
    setEditIdx(null);
    toast("Link updated!");
  };

  const handleCancelEdit = () => {
    setEditIdx(null);
  };

  const handleStart = () => {
    toast("Starting live search for selected links...");
  };

  return (
    <Card className="w-full max-w-3xl mx-auto mt-12">
      <CardHeader>
        <CardTitle>Live Search</CardTitle>
      </CardHeader>
      <CardContent>
        <form className="flex gap-4 mb-6" onSubmit={handleAdd}>
          <Input
            placeholder="Trade link URL"
            value={url}
            onChange={(e) => setUrl(e.target.value)}
            required
          />
          <Input
            placeholder="Description (optional)"
            value={description}
            onChange={(e) => setDescription(e.target.value)}
          />
          <Button type="submit">Add</Button>
        </form>
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead />
              <TableHead>League</TableHead>
              <TableHead>Search ID</TableHead>
              <TableHead>URL</TableHead>
              <TableHead>Description</TableHead>
              <TableHead>Status</TableHead>
              <TableHead>Actions</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {links.map((link, idx) => (
              <TableRow key={link.url}>
                <TableCell>
                  <Checkbox
                    checked={link.selected}
                    onCheckedChange={(checked) => handleSelect(idx, !!checked)}
                  />
                </TableCell>
                <TableCell>{link.league}</TableCell>
                <TableCell>{link.searchId}</TableCell>
                <TableCell>
                  {editIdx === idx ? (
                    <Input
                      value={editUrl}
                      onChange={(e) => setEditUrl(e.target.value)}
                      className="max-w-xs"
                    />
                  ) : (
                    <span className="truncate max-w-xs">{link.url}</span>
                  )}
                </TableCell>
                <TableCell>
                  {editIdx === idx ? (
                    <Input
                      value={editDescription}
                      onChange={(e) => setEditDescription(e.target.value)}
                      className="max-w-xs"
                    />
                  ) : (
                    <span className="truncate max-w-xs">{link.description}</span>
                  )}
                </TableCell>
                <TableCell>{link.status}</TableCell>
                {/* <TableCell>
                  {editIdx === idx ? (
                    <>
                      <Button size="sm" onClick={() => handleSaveEdit(idx)}>
                        Save
                      </Button>
                      <Button size="sm" variant="secondary" onClick={handleCancelEdit}>
                        Cancel
                      </Button>
                    </>
                  ) : (
                    <Button size="sm" variant="secondary" onClick={() => handleEdit(idx)}>
                      Edit
                    </Button>
                  )}
                </TableCell>*/}
                <TableCell>
                  {editIdx === idx ? (
                    <>
                      <Button size="sm" onClick={() => handleSaveEdit(idx)}>
                        Save
                      </Button>
                      <Button size="sm" variant="secondary" onClick={handleCancelEdit}>
                        Cancel
                      </Button>
                    </>
                  ) : (
                    <>
                      <Button size="sm" variant="secondary" onClick={() => handleEdit(idx)}>
                        Edit
                      </Button>
                      <Button
                        size="sm"
                        variant="destructive"
                        className="ml-2"
                        onClick={() => handleDelete(idx)}
                        aria-label="Delete"
                      >
                        <Trash2 className="w-4 h-4" />
                      </Button>
                    </>
                  )}
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </CardContent>
      <CardFooter>
        <Button className="w-full" onClick={handleStart}>
          Start Live Search
        </Button>
      </CardFooter>
    </Card>
  );
}
