import React, { useEffect, useState } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { loadConfig, saveConfig } from "@/services/configService";
import {
  Card,
  CardHeader,
  CardTitle,
  CardDescription,
  CardContent,
  CardFooter,
} from "@/components/ui/card";
import { Switch } from "@/components/ui/switch";
import { toast } from "sonner";
import { settings } from "../../wailsjs/go/models";
import Config = settings.Config;
import DefaultTradeLink = settings.DefaultTradeLink;

export default function Settings() {
  const [poeSessid, setPoeSessid] = useState("");
  const [accountName, setAccountName] = useState("");
  const [league, setLeague] = useState("");
  const [automationEnabled, setAutomationEnabled] = useState(false);
  const [delay, setDelay] = useState(1000);
  const [defaultTradeLinks, _] = useState<DefaultTradeLink[]>([]); // Placeholder for future use

  useEffect(() => {
    loadConfig().then((cfg: Config) => {
      setPoeSessid(cfg.poesessid || "");
      setAccountName(cfg.accountName || "");
      setLeague(cfg.league || "");
      setAutomationEnabled(cfg.automationEnabled || false);
      setDelay(cfg.delay || 1000);
    });
  }, []);

  const handleSave = async () => {
    const cfg = new settings.Config({
      poesessid: poeSessid,
      accountName,
      league,
      automationEnabled,
      delay,
      defaultTradeLinks,
    });
    await saveConfig(cfg);
    toast("Settings saved!");
  };

  return (
    <Card className="w-full max-w-2xl mx-auto mt-12">
      <CardHeader>
        <CardTitle>Settings</CardTitle>
        <CardDescription>
          Configure your Path of Exile automation tool settings below.
        </CardDescription>
        {/* Optional: Add a CardAction for extra actions */}
        {/* <CardAction>
          <Button variant="link">Help</Button>
        </CardAction> */}
      </CardHeader>
      <CardContent>
        <form
          id="settings-form"
          className="flex flex-col gap-6"
          onSubmit={(e) => {
            e.preventDefault();
            handleSave();
          }}
        >
          <div className="grid gap-2">
            <Label htmlFor="poe-sessid">POESESSID</Label>
            <Input
              id="poe-sessid"
              type="text"
              value={poeSessid}
              onChange={(e) => setPoeSessid(e.target.value)}
              placeholder="Enter your POESESSID"
              required
            />
          </div>
          <div className="grid gap-2">
            <Label htmlFor="account-name">Account Name</Label>
            <Input
              id="account-name"
              type="text"
              value={accountName}
              onChange={(e) => setAccountName(e.target.value)}
              placeholder="Enter your account name"
            />
          </div>
          <div className="grid gap-2">
            <Label htmlFor="league">League</Label>
            <Input
              id="league"
              type="text"
              value={league}
              onChange={(e) => setLeague(e.target.value)}
              placeholder="e.g. Affliction"
            />
          </div>
          <div className="flex items-center justify-between">
            <Label htmlFor="automation-enabled">Enable Automation</Label>
            <Switch
              id="automation-enabled"
              checked={automationEnabled}
              onCheckedChange={setAutomationEnabled}
            />
          </div>
          <div className="grid gap-2">
            <Label htmlFor="delay">Delay (ms)</Label>
            <Input
              id="delay"
              type="number"
              min={100}
              max={10000}
              value={delay}
              onChange={(e) => setDelay(Number(e.target.value))}
            />
          </div>
        </form>
      </CardContent>
      <CardFooter>
        <Button type="submit" form="settings-form" className="w-full mt-2">
          Save
        </Button>
      </CardFooter>
    </Card>
  );
}
