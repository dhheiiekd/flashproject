document.addEventListener("DOMContentLoaded", async () => {
    const urlParams = new URLSearchParams(window.location.search);
    const deckId = urlParams.get('deck_id') || '1';

    let cards = [];
    let currentIndex = 0;
    let isFlipped = false;
    let correctCount = 0;
    let isAnimating = false;

    const flashcard = document.getElementById('flashcard');
    const questionText = document.getElementById('questionText');
    const answerText = document.getElementById('answerText');
    const progressText = document.getElementById('progressText');
    const progressBarFill = document.getElementById('progressBarFill');
    const forgotBtn = document.getElementById('forgotBtn');
    const knewBtn = document.getElementById('knewBtn');
    const controlsArea = document.getElementById('controlsArea');
    const resultsScreen = document.getElementById('resultsScreen');

    try {
        // Используем твой рабочий эндпоинт для получения данных!
        const response = await fetch(`/api/cards?deck_id=${deckId}`);
        if (!response.ok) throw new Error("Ошибка при загрузке карт");

        cards = await response.json();

        if (!cards || cards.length === 0) {
            questionText.textContent = "В этой колоде пока нет карт для изучения.";
            return;
        }

        // Перемешиваем карты перед началом, чтобы было интереснее
        cards.sort(() => Math.random() - 0.5);
        renderCard();

    } catch (err) {
        console.error("Ошибка:", err);
        questionText.textContent = "Не удалось загрузить колоду.";
    }

    function renderCard() {
        if (currentIndex >= cards.length) {
            showResults();
            return;
        }

        isFlipped = false;
        isAnimating = false;

        forgotBtn.disabled = true;
        knewBtn.disabled = true;

        // Сбрасываем карту в 0 градусов перед показом нового вопроса
        gsap.set(flashcard, { rotationY: 0 });

        // Внимание: твоя структура использует card.front_text и card.back_text
        const currentCard = cards[currentIndex];
        questionText.textContent = currentCard.front_text || "Пустой вопрос";
        answerText.textContent = currentCard.back_text || "Пустой ответ";

        progressText.textContent = `Карточка ${currentIndex + 1} из ${cards.length}`;
        const progressPercent = (currentIndex / cards.length) * 100;
        progressBarFill.style.width = `${progressPercent}%`;
    }

    // Анимация вращения на 540 градусов
    flashcard.addEventListener('click', () => {
        if (isFlipped || isAnimating || cards.length === 0) return;

        isAnimating = true;

        gsap.to(flashcard, {
            rotationY: 540,
            duration: 0.8,
            ease: "back.out(1.1)",
            onComplete: () => {
                isFlipped = true;
                isAnimating = false;
                forgotBtn.disabled = false;
                knewBtn.disabled = false;
            }
        });
    });

    function handleVerdict(knew) {
        if (!isFlipped || isAnimating) return;
        if (knew) correctCount++;

        isAnimating = true;

        // Анимация улетания карты влево/вправо
        gsap.to(flashcard, {
            x: knew ? 300 : -300,
            opacity: 0,
            scale: 0.8,
            duration: 0.3,
            onComplete: () => {
                currentIndex++;
                gsap.set(flashcard, { x: 0, opacity: 1, scale: 1, rotationY: 0 });
                renderCard();
            }
        });
    }

    forgotBtn.addEventListener('click', () => handleVerdict(false));
    knewBtn.addEventListener('click', () => handleVerdict(true));

    function showResults() {
        progressBarFill.style.width = "100%";
        progressText.textContent = "Завершено";
        flashcard.style.display = 'none';
        controlsArea.style.display = 'none';

        resultsScreen.style.display = 'block';
        document.getElementById('scoreBadge').textContent = `${correctCount} / ${cards.length}`;
        
        const percent = (correctCount / cards.length) * 100;
        const subtitle = document.getElementById('resultsSubtitle');
        if (percent === 100) subtitle.textContent = "🔥 Идеально! Безупречные знания!";
        else if (percent >= 70) subtitle.textContent = "⚡ Отлично сработано! ";
        else subtitle.textContent = "Стоит повторить материал ещё раз.";

        // КОРРЕКТИРОВКА: Отправка JSON сессии на бэкенд в соответствии с твоей структурой данных
        fetch('/api/study-sessions', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
                deck_id: parseInt(deckId),
                score: correctCount,        // Твоё поле из структуры
                total_cards: cards.length   // Твоё поле из структуры
            })
        }).catch(err => console.error("Не удалось сохранить сессию:", err));
    }
});