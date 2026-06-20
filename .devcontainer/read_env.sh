#!/bin/bash
# Läs .env fil och skriv ut variabler (exkludera kommentarer)
grep -v '^#' .env | grep -v '^$'
