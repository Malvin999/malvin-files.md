// const input = document.getElementById('input-field');
// const chatContainer = document.getElementById('chat-container');
// const sendButton = document.getElementById('send-button');
// const messagesContainer = document.getElementById('messages');
// const commandPopup = document.getElementById('command-popup');
// let focusedCommandIndex = 0;

function sendMessage() {
    const msg = toMarkdown();
    if (msg === '') return;

    addMessage(msg, 'user');

    if (msg.startsWith('/')) {
        let cmd = {
            n: msg.substring(1),
            t: "cmd"
        }
        replyCmd(JSON.stringify(cmd));
    } else {
        reply(msg);
        // let update = await window.newUpdate(msg, null)
        // processResponse(await window.send(update));
    }

    clearInput();
    input.rows = 1;
    sendButton.disabled = true;
    // document.querySelector('#messages').scrollTop = messagesContainer.scrollHeight;
}

function processResponse(response) {
    response = JSON.parse(response);
    document.querySelectorAll('.bot').forEach(element => element.remove());
    document.querySelectorAll('.button-container').forEach(element => element.remove());
    response.Messages?.forEach(message => {
        const bubble = addMessage(message.Text, 'bot');
        document.querySelector('#messages').appendChild(bubble);
        if (message.Buttons && message.Buttons.length > 0) {
            attachKeyboard(message.Buttons)
        }
    });
    input.focus();
}

function attachKeyboard(buttons) {
    const buttonContainer = document.createElement('div');
    buttonContainer.classList.add('button-container');

    buttons.forEach((row) => {
        if (Array.isArray(row)) {
            const rowContainer = document.createElement('div');
            rowContainer.classList.add('button-row');

            row.forEach((btn) => {
                const button = document.createElement('button');
                button.innerText = btn.Name;
                button.classList.add('button'); // Add a class for styling
                button.onclick = async () => {
                    if (btn.Cmd.t === "iq") {
                        let search = '';
                        if (btn.Cmd.p?.length > 0) {
                            search = btn.Cmd.p[0];
                        }
                        openSearchModal(search);
                        return;
                    }
                    // let update = await window.newUpdate('', btn.Cmd)
                    // processResponse(await window.send(update));
                    replyCmd(JSON.stringify(btn.Cmd))
                };
                rowContainer.appendChild(button);
            });
            buttonContainer.appendChild(rowContainer);
        } else {
            if (row.Cmd.t === "iq") {
                return;
            }
            const button = document.createElement('button');
            button.innerText = row.Name;
            button.classList.add('button'); // Add a class for styling
            button.onclick = async () => {
                replyCmd(JSON.stringify(row.Cmd))
                // let update = await window.newUpdate('', row.Cmd)
                // processResponse(await window.send(update));
            };
            buttonContainer.appendChild(button);
        }
    });

    document.querySelector('#messages').appendChild(buttonContainer);
    const messagesContainer = document.querySelector('#messages');
    messagesContainer.scrollTop = messagesContainer.scrollHeight;
}

function addMessage(text, author) {
    const urlPattern = /(\bhttps?:\/\/[^\s]+)/g;
    text = text.replace(urlPattern, '<a href="$1">$1</a>');

    const messageBubble = document.createElement('div');
    messageBubble.classList.add('message', author);
    messageBubble.innerHTML = text.replace(/\n/g, '<br>');
    messagesContainer.appendChild(messageBubble);

    return messageBubble
}

const commands = [
    {command: '/today', display: '🏠 Today'},
    {command: '/files', display: '📄 Files'},
    {command: '/dirs', display: '🗂 Dirs'},
    {command: '/checklists', display: '☑️ Checklists'},
    {command: '/schedule', display: '📆 Schedule'},
    {command: '/stats', display: '📊 Stats'},
    {command: '/postpone', display: '🦥 Postpone'},
    {command: '/rename', display: '✏️ Rename'},
    {command: '/move', display: '➡️ Move'},
    {command: '/settings', display: '⚙️ Settings'},
    {command: '/help', display: '📕 Help'}
];

document.addEventListener('scroll', () => {
    input.focus();
});

// input.addEventListener('input', () => {
//     sendButton.disabled = toMarkdown() === '';
//
//     if (toMarkdown().startsWith('/')) {
//         showCommandPopup();
//     } else {
//         hideCommandPopup();
//     }
// });
//
// input.addEventListener('keydown', (event) => {
//     if (event.key === 'Enter' && !event.shiftKey) {
//         event.preventDefault();
//         if (toMarkdown().startsWith('/')) {
//             insertFocusedCommand();
//         } else {
//             sendMessage();
//         }
//     }
// });

// sendButton.addEventListener('click', sendMessage);

document.body.addEventListener('click', function (e) {
    if (e.target && e.target.nodeName == 'A' && e.target.href) {
        const url = e.target.href;
        if (
            !url.startsWith('http://#') &&
            !url.startsWith('https://#') &&
            !url.startsWith('file://') &&
            !url.startsWith('http://wails.localhost:')
        ) {
            e.preventDefault();
            window.runtime.BrowserOpenURL(url);
        }
    }
});

// window.addEventListener("focus", async () => {
//     input.focus();
// });

async function read(args) {
    let path = args[0];
    let fileHandle = await getFileHandle(path)
    let file = await fileHandle.getFile();

    return await file.text();
}

async function write(args) {
    let path = args[0];
    let content = args[1];

    let fileHandle = await getFileHandle(path, true);
    const writable = await fileHandle.createWritable();
    await writable.write(content);
    await writable.close();
}

async function exists(args) {
    let path = args[0]
    if (path === "") {
        return true
    }

    try {
        await getFileHandle(path);
        return true;
    } catch (error) {
        if (error.name === 'NotFoundError') {
            return false
        }
        // TODO better way to handle dirs
        if (error.name === 'TypeMismatchError') {
            return true;
        }
        console.log("EXISTS:", error, "PATH: ", path);
        throw error
    }
}

async function remove(args) {
    let path = args[0]
    let fileHandle = await getFileHandle(path);
    await fileHandle.remove()
}

async function rename(args) {
    let oldpath = args[0]
    let newpath = args[1]

    let content = await read([oldpath])
    await write([newpath, content])
    await remove([oldpath])
}


async function readDir(args) {
    let path = args[0];
    const dirs = path.split('/');
    let currentDirHandle = await getRootDirHandle();
    if (currentDirHandle === undefined) {
        console.error('PLS open dir');
        return [];
    }

    for (const dirName of dirs) {
        if (dirName) {
            currentDirHandle = await currentDirHandle.getDirectoryHandle(dirName, {create: false});
        }
    }

    const entries = [];

    for await (const [name, handle] of currentDirHandle.entries()) {
        if (handle.kind === 'directory') {
            entries.push({
                name: name,
                isDir: true,
                modTime: 0 // File System Access API doesn't provide modTime for directories
            });
        } else if (handle.kind === 'file') {
            const file = await handle.getFile();
            entries.push({
                name: name,
                isDir: false,
                modTime: file.lastModified // Returns timestamp in milliseconds
            });
        }
    }

    entries.sort((a, b) => a.name.toLowerCase().localeCompare(b.name.toLowerCase()));

    return entries;
}

async function mkdir(args) {
    let path = args[0];
    try {
        let currentDirHandle = await getRootDirHandle();
        await currentDirHandle.getDirectoryHandle(path, {create: true});
    } catch (error) {
        console.log(error);
        throw error;
    }
}

async function mkdirAll(args) {
    let path = args[0]
    const dirs = path.split('/');
    let currentDirHandle = await getRootDirHandle();
    for (const dirName of dirs) {
        if (dirName) {
            await mkdir([path])
        }
    }
}

// function receive(val) {
//     // processResponse(val)
// }

function createFileElement(fileName, isImage, markdownText, file = null) {
    const element = document.createElement('span');
    if (isImage && file) {
        element.className = 'md-image';

        // Create a small thumbnail image
        const img = document.createElement('img');
        img.src = URL.createObjectURL(file);
        img.style.width = '60px';
        img.style.height = '60px';
        img.style.objectFit = 'cover';
        img.style.borderRadius = '3px';
        img.style.verticalAlign = 'middle';
        img.title = fileName;

        element.appendChild(img);

        element.addEventListener('remove', () => {
            URL.revokeObjectURL(img.src);
        });
    } else if (isImage) {
        element.className = 'md-image';
        element.textContent = '🖼️';
        element.title = fileName;
    } else {
        element.className = 'md-link';
        element.textContent = '📎 ' + fileName;
    }
    element.contentEditable = false;
    element.dataset.markdown = markdownText;
    return element;
}

function wrapSelection(tagName) {
    const selection = window.getSelection();
    if (selection.rangeCount === 0 || selection.isCollapsed) return;

    const range = selection.getRangeAt(0);
    const selectedText = range.extractContents();

    const wrapper = document.createElement(tagName);
    wrapper.appendChild(selectedText);
    range.insertNode(wrapper);

    // Insert a zero-width space after the formatted element to break formatting
    const breakNode = document.createTextNode('\u200B'); // Zero-width space
    range.setStartAfter(wrapper);
    range.insertNode(breakNode);

    // Position cursor after the break node
    range.setStartAfter(breakNode);
    range.collapse(true);
    selection.removeAllRanges();
    selection.addRange(range);

}

function toMarkdown() {
    function processNode(node) {
        if (node.nodeType === Node.TEXT_NODE) {
            return node.textContent;
        } else if (node.nodeType === Node.ELEMENT_NODE) {
            // Check for markdown data first
            if (node.dataset && node.dataset.markdown) {
                return node.dataset.markdown;
            }

            const text = Array.from(node.childNodes).map(processNode).join('');
            const tagName = node.tagName.toLowerCase();

            switch (tagName) {
                case 'strong':
                case 'b':
                    return `**${text}**`;
                case 'em':
                case 'i':
                    return `*${text}*`;
                case 'br':
                    return '\n';
                case 'div':
                    return text + '\n';
                default:
                    return text;
            }
        }
        return '';
    }

    return Array.from(input.childNodes).map(processNode).join('').trim();
}

function clearInput() {
    input.innerHTML = '';
}

async function writeMediaFile(fileName, file) {
    try {
        const rootHandle = await getRootDirHandle();

        let mediaHandle;
        try {
            mediaHandle = await rootHandle.getDirectoryHandle('media');
        } catch {
            mediaHandle = await rootHandle.getDirectoryHandle('media', {create: true});
        }

        const fileHandle = await mediaHandle.getFileHandle(fileName, {create: true});
        const writable = await fileHandle.createWritable();
        await writable.write(file);
        await writable.close();

        const path = '/media/' + fileName;
        addMemFile(path, {
            isFile: true,
            path: path,
            imageUrl: await getImageUrl(fileHandle),
        });

        return fileHandle;
    } catch (error) {
        console.error('Error saving file:', error);
        return null;
    }
}

function generateSafeFileName(originalName) {
    const now = new Date();
    const timestamp = `${String(now.getDate()).padStart(2, '0')}.${String(now.getMonth() + 1).padStart(2, '0')}.${now.getFullYear()} ${String(now.getHours()).padStart(2, '0')}:${String(now.getMinutes()).padStart(2, '0')}`;
    return `${timestamp}-${originalName}`.replace(/[<>:"/\\|?*\s]/g, '-');
}

// Handle image/file pasting
// input.addEventListener('paste', async (event) => {
//     event.preventDefault();
//     const items = event.clipboardData.items;
//     let hasFile = false;
//     for (const item of items) {
//         if (item.kind === 'file') {
//             hasFile = true;
//
//             const file = item.getAsFile();
//             const fileName = generateSafeFileName(file.name);
//             const isImage = file.type.startsWith('image/');
//
//
//             const saved = await saveFile(fileName, file);
//             if (saved) {
//                 const markdownText = isImage
//                     ? `![${fileName}](media/${fileName})`
//                     : `[${fileName}](media/${fileName})`;
//
//                 const fileElement = createFileElement(fileName, isImage, markdownText, file);
//
//                 if (input.children.length === 1 && input.children[0].tagName === 'BR') {
//                     input.children[0].remove();
//                 }
//
//                 input.appendChild(fileElement);
//
//                 if (isImage) {
//                     const br1 = document.createElement('br');
//                     const br2 = document.createElement('br');
//
//                     input.appendChild(br1);
//                     input.appendChild(br2);
//
//                     const range = document.createRange();
//                     range.setStartAfter(br2);
//                     range.collapse(true);
//                     const selection = window.getSelection();
//                     selection.removeAllRanges();
//                     selection.addRange(range);
//                 }
//             } else {
//                 loadingSpan.textContent = '❌';
//                 loadingSpan.style.color = 'red';
//             }
//         }
//     }
//     if (!hasFile) {
//         const plainText = event.clipboardData.getData('text/plain');
//         if (plainText) {
//             const selection = window.getSelection();
//             if (selection.rangeCount > 0) {
//                 const range = selection.getRangeAt(0);
//                 range.deleteContents();
//
//                 const lines = plainText.split('\n');
//                 for (let i = 0; i < lines.length; i++) {
//                     if (lines[i]) {
//                         range.insertNode(document.createTextNode(lines[i]));
//                         range.collapse(false);
//                     }
//                     if (i < lines.length - 1) {
//                         const br = document.createElement('br');
//                         range.insertNode(br);
//                         range.collapse(false);
//                     }
//                 }
//             }
//         }
//     }
// });
//
// input.addEventListener('keydown', (e) => {
//     if (e.ctrlKey || e.metaKey) {
//         switch (e.key) {
//             case 'b':
//                 e.preventDefault();
//                 wrapSelection('strong');
//                 break;
//             case 'i':
//                 e.preventDefault();
//                 wrapSelection('em');
//                 break;
//         }
//     }
// });

function showCommandPopup() {
    const inputText = toMarkdown();

    const filteredCommands = commands.filter(cmd => cmd.command.startsWith(inputText));

    commandPopup.innerHTML = '';
    filteredCommands.forEach((cmd, index) => {
        const cmdItem = document.createElement('div');
        cmdItem.classList.add('command-item');
        cmdItem.textContent = cmd.display;
        cmdItem.setAttribute('data-index', index);

        cmdItem.onclick = () => {
            replyCmd(JSON.stringify({n: cmd.command.substring(1), t: "cmd"}))
            clearInput();
            hideCommandPopup();
        };

        commandPopup.appendChild(cmdItem);
    });

    if (filteredCommands.length === 0) {
        hideCommandPopup();
        return;
    }

    focusedCommandIndex = 0;
    updateFocusedCommand();

    commandPopup.classList.remove('hidden');
}

function updateFocusedCommand() {
    commandPopup.querySelectorAll('.command-item').forEach(item => {
        item.classList.remove('focused');
    });

    const commandItems = commandPopup.querySelectorAll('.command-item');
    if (commandItems[focusedCommandIndex]) {
        commandItems[focusedCommandIndex].classList.add('focused');
    }
}

function navigateCommands(direction) {
    const commandItems = commandPopup.querySelectorAll('.command-item');
    if (commandItems.length === 0) return;

    if (direction === 'up') {
        focusedCommandIndex = focusedCommandIndex === 0 ? commandItems.length - 1 : focusedCommandIndex - 1;
    } else if (direction === 'down') {
        focusedCommandIndex = focusedCommandIndex === commandItems.length - 1 ? 0 : focusedCommandIndex + 1;
    }

    updateFocusedCommand();
}

function insertFocusedCommand() {
    const commandItems = commandPopup.querySelectorAll('.command-item');
    const focusedCommand = commandItems[focusedCommandIndex];

    if (focusedCommand) {
        const commandText = focusedCommand.textContent.trim();
        const selectedCommand = commands.find(cmd => cmd.display === commandText);
        if (selectedCommand) {
            replyCmd(JSON.stringify({n: selectedCommand.command.substring(1), t: "cmd"}));
            clearInput();
            hideCommandPopup();
        }
    }
}

// input.addEventListener('keydown', (event) => {
//     const isCommandMode = toMarkdown().startsWith('/');
//     const popupVisible = !commandPopup.classList.contains('hidden');
//
//     if (event.key === 'Enter' && !event.shiftKey) {
//         event.preventDefault();
//         if (isCommandMode && popupVisible) {
//             insertFocusedCommand();
//         } else if (isCommandMode) {
//             sendMessage();
//         } else {
//             sendMessage();
//         }
//     } else if (isCommandMode && popupVisible) {
//         if (event.key === 'ArrowUp') {
//             event.preventDefault();
//             navigateCommands('up');
//         } else if (event.key === 'ArrowDown') {
//             event.preventDefault();
//             navigateCommands('down');
//         } else if (event.key === 'Escape') {
//             event.preventDefault();
//             hideCommandPopup();
//         }
//     }
// });

function hideCommandPopup() {
    commandPopup.classList.add('hidden');
    commandPopup.innerHTML = '';
    focusedCommandIndex = 0;
}