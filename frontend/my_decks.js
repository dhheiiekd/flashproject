document.addEventListener("DOMContentLoaded", loadDecks);

// Загрузка списка колод пользователя с бэкенда
async function loadDecks() {
    try {
        const response = await fetch('/api/decks/my');
        if (!response.ok) throw new Error("Ошибка загрузки");
        
        const decks = await response.json();
        const container = document.getElementById('decksList');
        container.innerHTML = '';

        if (!decks || decks.length === 0) {
            document.getElementById('noDecks').style.display = 'block';
            return;
        }

        // Рендерим каждую прямоугольную карточку
        decks.forEach(deck => {
            const card = document.createElement('div');
            card.className = 'deck-card';
            card.innerHTML = `
                <div>
                    <h3 class="deck-title">${deck.title}</h3>
                    <p class="deck-desc">${deck.description || 'Без описания'}</p>
                </div>
                <div class="deck-btn-group">
                    <a href="/decks/preview?deck_id=${deck.id}" class="btn">Открыть</a>
                    <button class="btn-delete" data-id="${deck.id}">🗑 Измельчить</button>
                </div>
            `;
            container.appendChild(card);
        });

        // Навешиваем обработчики событий удаления программно, без инлайн-атрибутов onclick
        initDeleteEvents();

    } catch (err) {
        console.error("Ошибка при получении списка колод:", err);
    }
}

// Привязка обработчиков клика для удаления колод
function initDeleteEvents() {
    const deleteButtons = document.querySelectorAll('.btn-delete');
    deleteButtons.forEach(button => {
        button.addEventListener('click', function() {
            const deckId = this.getAttribute('data-id');
            deleteDeck(deckId);
        });
    });
}

// Удаление выбранной колоды
async function deleteDeck(deckId) {
    if (!confirm("Вы уверены, что хотите уничтожить эту колоду и все её карты? Это действие необратимо!")) {
        return;
    }

    try {
        const response = await fetch(`/api/decks/delete?deck_id=${deckId}`, { method: 'DELETE' });
        if (response.ok) {
            loadDecks(); // Перезагружаем интерфейс после успешного удаления
        } else {
            alert("Не удалось удалить колоду.");
        }
    } catch (err) {
        console.error("Ошибка при попытке удаления колоды:", err);
    }
}