import { setWorldConstructor } from "@cucumber/cucumber";
import { PriorityWorld } from "../step-definitions/world";

// Tell cucumber to use our custom World class
setWorldConstructor(PriorityWorld);
