import type React from "react";
import { useEffect, useState } from "react";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { Card, CardHeader, CardTitle, CardContent, CardFooter } from "@/components/ui/card";
import {
  listTradeLinks,
  addTradeLink,
  updateTradeLinks,
  startLiveSearch,
  stopLiveSearch,
} from "@/services/liveSearchService";
import { livesearch } from "../../wailsjs/go/models";
import { toast } from "sonner";
import TradeLink = livesearch.TradeLink;
import { DataTable } from "@/live-search/data-table";
import { getColumns } from "@/live-search/columns";
import { EventsOn } from "../../wailsjs/runtime";

export default function LiveSearch() {
  const [url, setUrl] = useState("");
  const [description, setDescription] = useState("");
  const [links, setLinks] = useState<TradeLink[]>([]);
  const [editIdx, setEditIdx] = useState<number | null>(null);
  const [editUrl, setEditUrl] = useState("");
  const [editDescription, setEditDescription] = useState("");
  const [isLiveSearchRunning, setIsLiveSearchRunning] = useState(false);

  useEffect(() => {
    listTradeLinks().then((links) => {
      console.log("Fetched trade links", links);
      setLinks(links);
    });

    const off = EventsOn("linkStatusChanged", (link: TradeLink) => {
      console.log("Received linkStatusChanged event", link);
      // Actualiza el estado local según el enlace recibido
      setLinks((prev) => prev.map((l) => (l.searchId === link.searchId ? { ...l, ...link } : l)));
    });
    return () => {
      off();
    };
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

  const handleStart = async () => {
    setIsLiveSearchRunning(true);
    toast("Starting live search for selected links...");
    const updatedLinks = await startLiveSearch();
    console.log("Live search started, updated links:", updatedLinks);
    setLinks(updatedLinks);
    if (updatedLinks.some((link) => link.status === "auth_error")) {
      toast.error("Your POESESSID is invalid or expired. Please update it in settings.");
      setIsLiveSearchRunning(false);
    }
  };

  const handleStop = async () => {
    await stopLiveSearch();
    setIsLiveSearchRunning(false);
    toast("Live search stopped.");
    // Optionally, refresh links status
    setLinks(await listTradeLinks());
  };

  const handleSelect = async (idx: number, selected: boolean) => {
    const updated = links.map((l, i) => (i === idx ? { ...l, selected } : l));
    setLinks(updated);
    await updateTradeLinks(updated);
  };

  return (
    <Card className="max-w-3xl w-full mx-auto mt-12">
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
        <DataTable
          columns={getColumns({
            editIdx,
            editUrl,
            editDescription,
            setEditUrl,
            setEditDescription,
            handleEdit,
            handleSaveEdit,
            handleCancelEdit,
            handleDelete,
            handleSelect,
            data: links,
          })}
          data={links}
        />
        {/*<Table>
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
                    <Button size="sm" variant="secondary" onClick={() => handleEdit(idx)}>
                      Edit
                    </Button>
                  )}
                </TableCell>
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
        </Table>*/}
      </CardContent>
      <CardFooter>
        {/*<Button className="w-full" onClick={handleStart}>
          Start Live Search
        </Button>*/}
        {isLiveSearchRunning ? (
          <Button className="w-full" variant="destructive" onClick={handleStop}>
            Stop Live Search
          </Button>
        ) : (
          <Button className="w-full" onClick={handleStart}>
            Start Live Search
          </Button>
        )}
      </CardFooter>
    </Card>
  );
}
