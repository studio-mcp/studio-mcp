#!/usr/bin/env node

import { Studio } from './studio';

// Get command line arguments (excluding node and script path)
const argv = process.argv.slice(2);

// Start the server
Studio.serve(argv);
