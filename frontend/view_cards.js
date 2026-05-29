document.addEventListener("DOMContentLoaded", async () => {
    const urlParams = new URLSearchParams(window.location.search);
    const deckId = urlParams.get('deck_id') || '1';

    // Настраиваем кнопку возврата к вееру динамически
    const backToPreviewBtn = document.getElementById('backToPreview');
    if (backToPreviewBtn) {
        backToPreviewBtn.href = `/decks/preview?deck_id=${deckId}`;
    }

    try {
        // Запрашиваем карточки у нашего API эндпоинта
        const response = await fetch(`/api/cards?deck_id=${deckId}`);
        if (!response.ok) throw new Error("Ошибка загрузки");
        
        const cards = await response.json();
        const grid = document.getElementById('cardsGrid');

        if (!cards || cards.length === 0) {
            document.getElementById('noCardsMessage').style.display = 'block';
            return;
        }

        // Рендерим прямоугольные флип-карточки
        cards.forEach(card => {
            const wrapper = document.createElement('div');
            wrapper.className = 'flashcard-wrapper';
            wrapper.id = `card-wrapper-${card.id}`;
            
            wrapper.innerHTML = `
                <div class="flashcard-inner">
                    <div class="card-front">
                        <button class="delete-card-btn" title="Удалить карточку" data-id="${card.id}">🗑</button>
                        <div>${card.front_text}</div>
                    </div>
                    <div class="card-back">
                        <div>${card.back_text}</div>
                    </div>
                </div>
            `;

            // Логика переворота по тапу/клику
            wrapper.addEventListener('click', () => {
                wrapper.classList.toggle('flipped');
            });

            grid.appendChild(wrapper);
        });

        // Безопасное навешивание событий на кнопки удаления внутри сгенерированных карточек
        initDeleteHandlers();

    } catch (err) {
        console.error("Ошибка получения карточек колоды:", err);
        document.getElementById('cardsGrid').innerHTML = '<p style="color:red; text-align:center;">Не удалось загрузить карточки.</p>';
    }
});

// Навешивание листенеров для предотвращения всплытия (stopPropagation) и вызова удаления
function initDeleteHandlers() {
    const deleteButtons = document.querySelectorAll('.delete-card-btn');
    deleteButtons.forEach(btn => {
        btn.addEventListener('click', function(event) {
            event.stopPropagation(); // Важно: останавливаем переворот карточки при нажатии на корзину
            const cardId = this.getAttribute('data-id');
            deleteSingleCard(cardId);
        });
    });
}

// Функция удаления одной карточки
async function deleteSingleCard(cardId) {
    if (!confirm("Уничтожить эту карточку?")) return;

    try {
        const response = await fetch(`/api/cards/delete?card_id=${cardId}`, { method: 'DELETE' });
        if (response.ok) {
            // Удаляем элемент из DOM дерева без перезагрузки страницы
            const cardElement = document.getElementById(`card-wrapper-${cardId}`);
            if (cardElement) cardElement.remove();
            
            // Если карточек внутри сетки больше не осталось, включаем сообщение-заглушку
            const grid = document.getElementById('cardsGrid');
            if (grid.children.length === 0) {
                document.getElementById('noCardsMessage').style.display = 'block';
            }
        } else {
            alert("Не удалось удалить карточку.");
        }
    } catch (err) {
        console.error("Ошибка при попытке удаления карточки:", err);
    }
}