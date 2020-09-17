"use strict";
//

//
export const output = document.getElementById("messages");
export const el_1 = document.getElementById("channel-list");
export const el_2 = document.getElementById("server-name");
export const el_3 = document.getElementById("me");
export const el_uonline = document.body.children[4];
export const el_input = document.getElementById("input");
export const context = new AudioContext();
export const audio_buffer_size = 4096;

export const msg_processors = [
    ["/shrug", "¯\\_(ツ)_/¯"],
    ["/tableflip", "(╯°□°）╯︵ ┻━┻"],
    ["/unflip", "┬─┬ ノ( ゜-゜ノ)"],
    ["/shobon","(´・ω・`)"],
];

export function getSettingsSelection(f, s) {
    return document.querySelector(`x-settings[data-s-for="${f}"] [data-s-section="${s}"] x-selection`);
}
