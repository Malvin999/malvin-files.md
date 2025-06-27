// Global variables
let messages = [];
let chatContainer;
let messageInput;
const CHAT_FILENAME = 'Chat.txt';



function parseFileContent(content) {
    const lines = content.split('\n');
    const messages = [];
    let currentDate = null;

    for (const line of lines) {
        const trimmedLine = line.trim();
        if (!trimmedLine) continue;

        // Check if line is a date header (starts with ####)
        if (trimmedLine.startsWith('####')) {
            currentDate = trimmedLine.replace(/^#+\s*/, '').trim();
            continue;
        }

        // Check if line is a timestamped message (starts with backtick)
        const timeMatch = trimmedLine.match(/^`(\d{2}:\d{2})`\s*(.*)$/);
        if (timeMatch) {
            const [, timestamp, text] = timeMatch;

            if (text.trim()) {
                messages.push({
                    id: Date.now() + Math.random(), // Generate unique ID
                    text: text.trim(),
                    timestamp: timestamp,
                    date: currentDate || new Date().toDateString()
                });
            }
        }
    }

    return messages;
}

function formatFileContent(messages) {
    if (messages.length === 0) return '';

    // Group messages by date
    const messagesByDate = {};
    messages.forEach(msg => {
        const date = msg.date || new Date().toDateString();
        if (!messagesByDate[date]) {
            messagesByDate[date] = [];
        }
        messagesByDate[date].push(msg);
    });

    let content = '';
    Object.entries(messagesByDate).forEach(([date, msgs]) => {
        if (content) content += '\n';
        content += `#### ${date}\n`;
        msgs.forEach(msg => {
            content += `\`${msg.timestamp}\` ${msg.text}\n`;
        });
    });

    return content;
}

async function loadData() {
    try {
        const file = await ((await getFileHandle(CHAT_FILENAME)).getFile());
        const content = await file.text();

        // Parse the content and load messages
        messages = parseFileContent(content);

        console.log(`Loaded ${messages.length} messages from ${CHAT_FILENAME}`);
    } catch (error) {
        console.error('Error loading data:', error);
        // Initialize with empty data if file doesn't exist or can't be read
        messages = [];
    }
}

async function saveData() {
    try {
        // For now, just save the current file's messages
        // You can extend this to save all files
        const content = formatFileContent(sidebarFiles[currentFile]);

        // You'll need to implement the file writing part
        // This is a placeholder for your file system API
        console.log('Would save to file:', content);

        // Example of what the save might look like:
        // const fileHandle = await getFileHandle(CHAT_FILENAME);
        // await fileHandle.write(content);

    } catch (error) {
        console.error('Error saving data:', error);
    }
}

function initChat() {
    chatContainer = document.getElementById('chat');
    messageInput = document.getElementById('chat-input');

    messageInput.addEventListener('keydown', function(e) {
        if (e.key === 'Enter' && !e.shiftKey) {
            e.preventDefault();
            handleSend();
        }
    });

    loadData().then(() => {
        renderMessages();
    });
}

async function handleSend() {
    const text = messageInput.value.trim();
    if (!text) return;

    // const now = new Date();
    // const note = {
    //     id: Date.now(),
    //     text: text,
    //     timestamp: now.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' }),
    //     date: now.toLocaleDateString('en-GB', {
    //         day: 'numeric',
    //         month: 'long',
    //         weekday: 'long'
    //     })
    // };
    //
    // sidebarFiles[currentFile].push(note);

    messageInput.value = '';
    // saveData();

    reply(text);
    await loadData();
    renderMessages();
    scrollToBottom();

    // Notify sidebar of changes (commented for future implementation)
    // updateSidebar();
}

function renderMessages() {
    if (messages.length === 0) {
        chatContainer.innerHTML = `
            <div class="empty-state">
                <svg class="empty-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor">
                    <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/>
                    <polyline points="14,2 14,8 20,8"/>
                    <line x1="16" y1="13" x2="8" y2="13"/>
                    <line x1="16" y1="17" x2="8" y2="17"/>
                    <polyline points="10,9 9,9 8,9"/>
                </svg>
                <div class="empty-title">No notes found</div>
                <div class="empty-desc">Start typing to add your first note</div>
            </div>
        `;
        return;
    }

    chatContainer.innerHTML = messages.map(note => `
        <div class="message" data-note-id="${note.id}">
            <div class="message-content" 
                 contenteditable="true" 
                 data-note-id="${note.id}"
                 spellcheck="false">${escapeHtml(note.text)}</div>
            <div class="message-footer">
                <span class="message-time">${note.timestamp}</span>
                <div class="message-actions">
                    <button class="action-btn delete-btn" data-note-id="${note.id}">
                        🗑️
                        <span class="btn-label">Delete</span>
                    </button>
                </div>
            </div>
        </div>
    `).join('');

    attachEventListeners();
}



function attachEventListeners() {
    // Add event listeners for editing message content
    chatContainer.querySelectorAll('.message-content[contenteditable]').forEach(element => {
        element.addEventListener('blur', function(e) {
            saveEdit(e.target.dataset.noteId, e.target.textContent);
            e.target.classList.remove('editing');
        });

        element.addEventListener('focus', function(e) {
            e.target.classList.add('editing');
        });

        element.addEventListener('keydown', function(e) {
            if (e.key === 'Enter' && !e.shiftKey) {
                e.preventDefault();
                e.target.blur();
            }
            if (e.key === 'Escape') {
                e.target.textContent = messages.find(n => n.id == e.target.dataset.noteId).text;
                e.target.blur();
            }
        });
    });

    // Add delete button listeners
    chatContainer.querySelectorAll('.delete-btn').forEach(btn => {
        btn.addEventListener('click', function(e) {
            e.stopPropagation();
            deleteNote(btn.dataset.noteId);
        });
    });
}



function saveEdit(noteId, newText) {
    const note = messages.find(n => n.id == noteId);
    if (note && newText.trim() !== '') {
        note.text = newText.trim();
        saveData();
    }
}

function deleteNote(noteId) {
    messages = messages.filter(n => n.id != noteId);
    saveData();
    renderMessages();
}

function scrollToBottom() {
    setTimeout(function() {
        chatContainer.scrollTop = chatContainer.scrollHeight;
    }, 100);
}

function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

// Initialize when page loads
document.addEventListener('DOMContentLoaded', init);

const chatInput = document.getElementById('chat-input');
function autoResize() {
    chatInput.style.height = 'auto';
    chatInput.style.height = Math.min(chatInput.scrollHeight, 250) + 'px';
}
// Add event listener for input changes
chatInput.addEventListener('input', autoResize);
// Initial resize to set proper height
autoResize();