import Logo from "@/assets/images/logo-universal.png";
import { ModeToggle } from "@/components/mode-toggle";
import { AspectRatio } from "@/components/ui/aspect-ratio";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import type React from "react";
import { useState } from "react";
import { toast } from "sonner";
import { Greet } from "../../wailsjs/go/backend/App";

const HelloWorld: React.FC = () => {
  const [resultText, setResultText] = useState("Please enter your name below");
  const [name, setName] = useState("");
  const updateName = (e: React.ChangeEvent<HTMLInputElement>) => setName(e.target.value);
  const updateResultText = (result: string) => setResultText(result);

  function greet() {
    Greet(name)
      .then(updateResultText)
      .then(() => toast("Greeted!"));
  }
  return (
    <Card className="p-4 mx-auto">
      <div className="flex align-end justify-end">
        <ModeToggle />
      </div>
      <div className="flex align-center justify-center mb-12">
        <AspectRatio ratio={16 / 9}>
          <img src={Logo} alt="Logo" />
        </AspectRatio>
      </div>
      <div className="text-md font-bold text-center">{resultText}</div>
      <Input
        id="name"
        onChange={updateName}
        autoComplete="off"
        name="input"
        type="text"
        placeholder="Enter your name"
        className="w-[20rem]"
      />
      <Button variant="outline" onClick={greet}>
        Greet
      </Button>
    </Card>
  );
};

export default HelloWorld;
