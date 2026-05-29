document.addEventListener("DOMContentLoaded", () => {
    let currentCategoryId = ""; // Храним выбранную категорию (пусто - значит все темы)

    const searchInput = document.getElementById('searchInput');
    const container = document.getElementById('decksList');
    const noDecksMsg = document.getElementById('noDecks');
    const categoryTagsContainer = document.getElementById('categoryTags');

    const urlParams = new URLSearchParams(window.location.search);
    const fromGroupId = urlParams.get('from_group');

    // 1. Загрузка категорий с бэкенда и генерация кнопок-тегов
    async function loadCategoryFilters() {
        try {
            const response = await fetch('/api/categories');
            if (!response.ok) return;
            const categories = await response.json();

            categories.forEach(cat => {
                const btn = document.createElement('button');
                btn.className = 'tag-btn';
                btn.textContent = cat.name;
                btn.setAttribute('data-id', cat.id);
                
                // Обработчик клика по категории
                btn.addEventListener('click', (e) => {
                    // Переключаем класс active у кнопок
                    document.querySelectorAll('.tag-btn').forEach(b => b.classList.remove('active'));
                    btn.classList.add('active');

                    // Запоминаем ID и отправляем запрос к БД
                    currentCategoryId = btn.getAttribute('data-id');
                    loadPublicDecks(); 
                });

                categoryTagsContainer.appendChild(btn);
            });
        } catch (err) {
            console.error("Не удалось загрузить теги категорий:", err);
        }
    }

    // Обработчик клика для самой первой дефолтной кнопки "Все подряд"
    const allBtn = categoryTagsContainer.querySelector('.tag-btn');
    if (allBtn) {
        allBtn.addEventListener('click', () => {
            document.querySelectorAll('.tag-btn').forEach(b => b.classList.remove('active'));
            allBtn.classList.add('active');
            currentCategoryId = "";
            loadPublicDecks();
        });
    }

    // 2. Функция загрузки колод с учетом живого поиска и фильтра по категории
    async function loadPublicDecks() {
        try {
            const query = searchInput ? searchInput.value.trim() : "";
            
            // Собираем динамический URL для бэкенда с учетом фильтра и поиска
            let url = `/api/decks/search?query=${encodeURIComponent(query)}`;
            if (currentCategoryId) {
                url += `&category_id=${currentCategoryId}`;
            }

            const response = await fetch(url);
            if (!response.ok) throw new Error("Ошибка при получении публичных колод");

            const decks = await response.json();
            renderDecks(decks);
        } catch (err) {
            console.error("Не удалось загрузить глобальную библиотеку:", err);
            if (container) {
                container.innerHTML = `<p class="no-decks-message" style="display:block; color:#ff6b6b;">Ошибка сервера при загрузке данных</p>`;
            }
        }
    }

    // 3. Функция генерации карточек в HTML
    function renderDecks(decks) {
        if (!container) return;
        container.innerHTML = '';

        if (!decks || decks.length === 0) {
            if (noDecksMsg) noDecksMsg.style.display = 'block';
            return;
        }

        if (noDecksMsg) noDecksMsg.style.display = 'none';

        decks.forEach(deck => {
            const card = document.createElement('div');
            card.className = 'deck-card';

            let actionButtonHtml = '';
            if (fromGroupId) {
                actionButtonHtml = `<button class="btn attach-btn" data-id="${deck.id}" style="background-color: #28a745; border: none; cursor: pointer; color: white; width: 100%;">📌 Прикрепить к клану</button>`;
            } else {
                actionButtonHtml = `<a href="/decks/play?deck_id=${deck.id}" class="btn">🚀 Пройти колоду</a>`;
            }

            card.innerHTML = `
                <div>
                    <h3 class="deck-title">${deck.title}</h3>
                    <p class="deck-desc">${deck.description || 'Без описания'}</p>
                </div>
                <div class="deck-btn-group" style="margin-top: 15px;">
                    ${actionButtonHtml}
                </div>
            `;
            container.appendChild(card);
        });

        if (fromGroupId) {
            document.querySelectorAll('.attach-btn').forEach(btn => {
                btn.addEventListener('click', async (e) => {
                    const deckId = e.target.getAttribute('data-id');
                    await attachDeckToGroup(fromGroupId, deckId);
                });
            });
        }
    }

    // 4. Функция прикрепления колоды к группе
    async function attachDeckToGroup(groupId, deckId) {
        const formData = new FormData();
        formData.append('group_id', groupId);
        formData.append('deck_id', deckId);

        try {
            const res = await fetch('/api/groups/decks/add', { 
                method: 'POST', 
                body: formData 
            });

            if (res.ok) {
                alert("Колода успешно добавлена в кузницу твоего клана!");
                window.location.href = `/decks/group-details?id=${groupId}`;
            } else {
                const errText = await res.text();
                alert("Не удалось прикрепить колоду: " + errText);
            }
        } catch (err) {
            console.error(err);
            alert("Ошибка соединения с сервером кузницы.");
        }
    }

    // 5. Живой поиск по ходу ввода букв
    if (searchInput) {
        searchInput.addEventListener('input', () => {
            loadPublicDecks(); // Просто перевызываем загрузку, она сама считает значение из инпута
        });
    }

    // Стартуем загрузку интерфейса
    loadCategoryFilters();
    loadPublicDecks();
});