package patch

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
	"time"

	"signature-menu-backend/internal/store"

	"golang.org/x/crypto/bcrypt"
)

const (
	mockUsername = "mock_data"
	mockPassword = "mock123456"
)

type ingredientSeed struct {
	name   string
	amount string
	unit   string
}

type stepSeed struct {
	title   string
	minutes int
}

type recipeSeed struct {
	name        string
	description string
	method      string
	tags        []string
	ingredients []ingredientSeed
	steps       []stepSeed
	minutes     int
	servings    int
	priceRange  string
	difficulty  int
	proficiency int
}

func runMockData(dataStore *store.Store) error {
	user, err := ensureMockUser(dataStore)
	if err != nil {
		return err
	}

	for _, recipe := range dataStore.ListRecipes(user.ID) {
		if err := dataStore.DeleteRecipe(user.ID, recipe.ID); err != nil {
			return err
		}
	}

	random := rand.New(rand.NewSource(time.Now().UnixNano()))
	seeds := mockRecipeSeeds()
	random.Shuffle(len(seeds), func(i int, j int) {
		seeds[i], seeds[j] = seeds[j], seeds[i]
	})

	count := 50
	if len(seeds) < count {
		count = len(seeds)
	}

	for index := 0; index < count; index++ {
		input := toRecipeMutation(seeds[index], random)
		if _, err := dataStore.CreateRecipe(user.ID, input); err != nil {
			return fmt.Errorf("create mock recipe %q: %w", seeds[index].name, err)
		}
	}

	log.Printf("mock data patch done: user=%s password=%s recipes=%d", mockUsername, mockPassword, count)
	return nil
}

func ensureMockUser(dataStore *store.Store) (store.User, error) {
	user, err := dataStore.FindUserByUsername(mockUsername)
	if err == nil {
		return user, nil
	}
	if !errors.Is(err, store.ErrNotFound) {
		return store.User{}, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(mockPassword), bcrypt.DefaultCost)
	if err != nil {
		return store.User{}, err
	}
	return dataStore.CreateUser(mockUsername, string(hash), "模拟菜谱")
}

func toRecipeMutation(seed recipeSeed, random *rand.Rand) store.RecipeMutation {
	ingredients := make([]store.IngredientMutation, 0, len(seed.ingredients))
	for _, ingredient := range seed.ingredients {
		ingredients = append(ingredients, store.IngredientMutation{
			Name:   ingredient.name,
			Amount: ingredient.amount,
			Unit:   ingredient.unit,
		})
	}

	steps := make([]store.StepMutation, 0, len(seed.steps))
	for index, step := range seed.steps {
		steps = append(steps, store.StepMutation{
			StepOrder:        index + 1,
			Title:            step.title,
			EstimatedMinutes: step.minutes,
		})
	}

	difficulty := seed.difficulty
	if difficulty <= 0 {
		difficulty = random.Intn(3) + 2
	}
	proficiency := seed.proficiency
	if proficiency <= 0 {
		proficiency = random.Intn(3) + 3
	}

	return store.RecipeMutation{
		Name:             seed.name,
		Description:      seed.description,
		CookingMethod:    seed.method,
		ServingCount:     seed.servings,
		EstimatedMinutes: seed.minutes,
		Difficulty:       difficulty,
		IsAvailable:      random.Intn(100) < 82,
		TasteTags:        seed.tags,
		Proficiency:      proficiency,
		PriceRange:       seed.priceRange,
		CookedCount:      random.Intn(18),
		PrivateNote:      "模拟数据，可直接编辑成自己的做法。",
		Ingredients:      ingredients,
		Steps:            steps,
	}
}

func mockRecipeSeeds() []recipeSeed {
	return []recipeSeed{
		{
			name: "番茄牛腩", description: "酸甜开胃，汤汁拌饭很稳。", method: "炖", tags: []string{"家常", "咸鲜", "下饭"}, minutes: 90, servings: 3, priceRange: "30-50元", difficulty: 3, proficiency: 4,
			ingredients: []ingredientSeed{{"牛腩", "500", "g"}, {"番茄", "3", "个"}, {"土豆", "1", "个"}, {"洋葱", "半", "个"}, {"姜片", "4", "片"}},
			steps:       []stepSeed{{"牛腩冷水下锅焯出血沫", 10}, {"番茄炒软出汁", 8}, {"加入牛腩和热水小火炖煮", 60}, {"放土豆收汁调味", 12}},
		},
		{
			name: "蒜蓉虾", description: "蒜香足，十几分钟能上桌。", method: "蒸", tags: []string{"家常", "蒜香", "鲜香"}, minutes: 18, servings: 2, priceRange: "25-45元", difficulty: 2, proficiency: 4,
			ingredients: []ingredientSeed{{"鲜虾", "400", "g"}, {"蒜", "1", "头"}, {"粉丝", "1", "把"}, {"生抽", "2", "勺"}, {"葱花", "适量", ""}},
			steps:       []stepSeed{{"粉丝泡软铺盘", 5}, {"虾开背去虾线", 5}, {"炒香蒜蓉调味铺在虾上", 4}, {"上锅蒸熟撒葱花", 4}},
		},
		{
			name: "青椒炒蛋", description: "最快手的一道，香辣下饭。", method: "炒", tags: []string{"家常", "下饭", "快手"}, minutes: 10, servings: 2, priceRange: "10元以内", difficulty: 1, proficiency: 5,
			ingredients: []ingredientSeed{{"鸡蛋", "3", "个"}, {"青椒", "2", "个"}, {"盐", "少许", ""}, {"生抽", "1", "勺"}},
			steps:       []stepSeed{{"鸡蛋打散炒至定型盛出", 3}, {"青椒切块下锅炒香", 3}, {"倒回鸡蛋调味翻匀", 4}},
		},
		{
			name: "冬瓜丸子汤", description: "清爽不腻，适合两三个人。", method: "汤", tags: []string{"清淡", "鲜美", "家常"}, minutes: 30, servings: 3, priceRange: "15-25元", difficulty: 2, proficiency: 4,
			ingredients: []ingredientSeed{{"冬瓜", "400", "g"}, {"猪肉馅", "250", "g"}, {"姜末", "少许", ""}, {"香菜", "少许", ""}},
			steps:       []stepSeed{{"肉馅加姜末和调料搅上劲", 8}, {"冬瓜切片煮到半透明", 8}, {"挤入丸子煮熟", 10}, {"撒香菜出锅", 2}},
		},
		{
			name: "葱油鸡", description: "鸡肉嫩，葱油香，冷吃热吃都行。", method: "蒸", tags: []string{"咸鲜", "葱香", "聚餐"}, minutes: 35, servings: 3, priceRange: "25-40元", difficulty: 3, proficiency: 4,
			ingredients: []ingredientSeed{{"鸡腿", "3", "只"}, {"小葱", "1", "把"}, {"姜片", "5", "片"}, {"生抽", "2", "勺"}},
			steps:       []stepSeed{{"鸡腿加姜葱蒸熟", 25}, {"撕成大块装盘", 4}, {"热油激香葱花", 3}, {"淋入调味汁", 3}},
		},
		{
			name: "红烧排骨", description: "甜咸适中，适合周末慢慢烧。", method: "烧", tags: []string{"家常", "红烧", "下饭"}, minutes: 60, servings: 3, priceRange: "35-55元", difficulty: 3, proficiency: 4,
			ingredients: []ingredientSeed{{"排骨", "600", "g"}, {"冰糖", "20", "g"}, {"姜片", "4", "片"}, {"料酒", "2", "勺"}, {"生抽", "2", "勺"}},
			steps:       []stepSeed{{"排骨焯水洗净", 10}, {"炒糖色后下排骨翻匀", 8}, {"加调料和热水焖烧", 35}, {"大火收汁", 7}},
		},
		{
			name: "鱼香肉丝", description: "酸甜微辣，米饭杀手。", method: "炒", tags: []string{"酸甜", "微辣", "下饭"}, minutes: 25, servings: 2, priceRange: "15-25元", difficulty: 3, proficiency: 3,
			ingredients: []ingredientSeed{{"里脊肉", "250", "g"}, {"木耳", "1", "把"}, {"胡萝卜", "半", "根"}, {"青椒", "1", "个"}, {"郫县豆瓣", "1", "勺"}},
			steps:       []stepSeed{{"肉丝腌制上浆", 8}, {"配菜切丝", 5}, {"炒香豆瓣和肉丝", 6}, {"倒入鱼香汁快速翻炒", 6}},
		},
		{
			name: "宫保鸡丁", description: "鸡丁嫩，花生脆，甜辣平衡。", method: "炒", tags: []string{"微辣", "咸甜", "下饭"}, minutes: 25, servings: 2, priceRange: "18-30元", difficulty: 3, proficiency: 3,
			ingredients: []ingredientSeed{{"鸡腿肉", "300", "g"}, {"花生米", "50", "g"}, {"干辣椒", "6", "个"}, {"黄瓜", "半", "根"}, {"葱白", "适量", ""}},
			steps:       []stepSeed{{"鸡肉切丁腌制", 8}, {"调宫保汁", 3}, {"炒香辣椒和鸡丁", 8}, {"加花生和汁翻匀", 6}},
		},
		{
			name: "可乐鸡翅", description: "孩子也爱吃，甜香入味。", method: "烧", tags: []string{"甜咸", "家常", "宴客"}, minutes: 35, servings: 3, priceRange: "20-35元", difficulty: 2, proficiency: 4,
			ingredients: []ingredientSeed{{"鸡翅中", "10", "个"}, {"可乐", "1", "听"}, {"姜片", "3", "片"}, {"生抽", "2", "勺"}},
			steps:       []stepSeed{{"鸡翅划刀焯水", 8}, {"煎到两面微黄", 8}, {"加可乐和生抽焖煮", 15}, {"收浓汤汁", 4}},
		},
		{
			name: "麻婆豆腐", description: "麻辣鲜香，热乎乎最下饭。", method: "烧", tags: []string{"麻辣", "下饭", "川味"}, minutes: 18, servings: 2, priceRange: "10-18元", difficulty: 2, proficiency: 4,
			ingredients: []ingredientSeed{{"嫩豆腐", "1", "盒"}, {"牛肉末", "80", "g"}, {"豆瓣酱", "1", "勺"}, {"花椒粉", "少许", ""}},
			steps:       []stepSeed{{"豆腐切块焯水", 4}, {"炒香肉末和豆瓣", 5}, {"加水烧豆腐入味", 6}, {"勾芡撒花椒粉", 3}},
		},
		{
			name: "糖醋里脊", description: "外酥里嫩，酸甜亮汁。", method: "炸", tags: []string{"酸甜", "宴客", "香脆"}, minutes: 35, servings: 3, priceRange: "20-35元", difficulty: 4, proficiency: 3,
			ingredients: []ingredientSeed{{"里脊肉", "350", "g"}, {"淀粉", "80", "g"}, {"番茄酱", "3", "勺"}, {"白醋", "1", "勺"}, {"白糖", "2", "勺"}},
			steps:       []stepSeed{{"里脊切条腌制", 8}, {"挂糊炸至金黄", 14}, {"熬糖醋汁", 5}, {"倒入里脊翻裹", 4}},
		},
		{
			name: "酸菜鱼", description: "酸辣开胃，鱼片嫩滑。", method: "煮", tags: []string{"酸辣", "聚餐", "鲜香"}, minutes: 45, servings: 4, priceRange: "35-60元", difficulty: 4, proficiency: 3,
			ingredients: []ingredientSeed{{"黑鱼片", "500", "g"}, {"酸菜", "250", "g"}, {"豆芽", "200", "g"}, {"泡椒", "适量", ""}},
			steps:       []stepSeed{{"鱼片加盐和淀粉腌制", 10}, {"酸菜炒香加水煮汤", 12}, {"下配菜煮熟", 8}, {"滑入鱼片烫熟", 5}, {"热油泼香", 4}},
		},
		{
			name: "清蒸鲈鱼", description: "鲜嫩清爽，宴客不费力。", method: "蒸", tags: []string{"清淡", "鲜美", "宴客"}, minutes: 20, servings: 3, priceRange: "35-55元", difficulty: 2, proficiency: 4,
			ingredients: []ingredientSeed{{"鲈鱼", "1", "条"}, {"葱丝", "适量", ""}, {"姜片", "适量", ""}, {"蒸鱼豉油", "2", "勺"}},
			steps:       []stepSeed{{"鱼处理干净打花刀", 5}, {"铺姜片上锅蒸", 10}, {"倒掉蒸汁铺葱丝", 2}, {"淋豉油热油", 3}},
		},
		{
			name: "蒜泥白肉", description: "肥而不腻，蒜香微辣。", method: "拌", tags: []string{"蒜香", "微辣", "凉菜"}, minutes: 35, servings: 3, priceRange: "25-40元", difficulty: 3, proficiency: 3,
			ingredients: []ingredientSeed{{"五花肉", "400", "g"}, {"黄瓜", "1", "根"}, {"蒜", "半", "头"}, {"辣椒油", "2", "勺"}},
			steps:       []stepSeed{{"五花肉冷水下锅煮熟", 25}, {"放凉切薄片", 5}, {"调蒜泥料汁", 3}, {"铺黄瓜拌匀", 2}},
		},
		{
			name: "凉拌黄瓜", description: "清爽解腻，五分钟搞定。", method: "拌", tags: []string{"清淡", "凉菜", "快手"}, minutes: 8, servings: 2, priceRange: "10元以内", difficulty: 1, proficiency: 5,
			ingredients: []ingredientSeed{{"黄瓜", "2", "根"}, {"蒜末", "适量", ""}, {"香醋", "1", "勺"}, {"生抽", "1", "勺"}},
			steps:       []stepSeed{{"黄瓜拍裂切段", 3}, {"蒜末和调料拌匀", 3}, {"静置入味", 2}},
		},
		{
			name: "地三鲜", description: "软糯咸香，东北家常味。", method: "烧", tags: []string{"家常", "咸鲜", "素菜"}, minutes: 30, servings: 3, priceRange: "12-20元", difficulty: 3, proficiency: 3,
			ingredients: []ingredientSeed{{"土豆", "2", "个"}, {"茄子", "1", "根"}, {"青椒", "1", "个"}, {"蒜末", "适量", ""}},
			steps:       []stepSeed{{"土豆和茄子煎炸至软", 15}, {"青椒断生", 3}, {"炒香蒜末倒入酱汁", 4}, {"三鲜回锅裹汁", 5}},
		},
		{
			name: "干煸四季豆", description: "豆角焦香，肉末提味。", method: "炒", tags: []string{"微辣", "下饭", "家常"}, minutes: 25, servings: 2, priceRange: "12-22元", difficulty: 3, proficiency: 3,
			ingredients: []ingredientSeed{{"四季豆", "350", "g"}, {"肉末", "80", "g"}, {"干辣椒", "5", "个"}, {"蒜末", "适量", ""}},
			steps:       []stepSeed{{"四季豆煸到起皱", 12}, {"炒香肉末和辣椒", 5}, {"回锅调味翻炒", 6}},
		},
		{
			name: "蒜蓉西兰花", description: "清爽脆嫩，配肉菜很合适。", method: "炒", tags: []string{"清淡", "素菜", "蒜香"}, minutes: 12, servings: 2, priceRange: "10-18元", difficulty: 1, proficiency: 5,
			ingredients: []ingredientSeed{{"西兰花", "1", "颗"}, {"蒜", "4", "瓣"}, {"蚝油", "1", "勺"}, {"盐", "少许", ""}},
			steps:       []stepSeed{{"西兰花焯水", 4}, {"蒜末炒香", 2}, {"下西兰花调味快炒", 4}},
		},
		{
			name: "西红柿炒鸡蛋", description: "酸甜软嫩，家里常备菜。", method: "炒", tags: []string{"家常", "酸甜", "快手"}, minutes: 12, servings: 2, priceRange: "10元以内", difficulty: 1, proficiency: 5,
			ingredients: []ingredientSeed{{"西红柿", "2", "个"}, {"鸡蛋", "3", "个"}, {"葱花", "少许", ""}, {"糖", "少许", ""}},
			steps:       []stepSeed{{"鸡蛋炒熟盛出", 3}, {"西红柿炒出汁", 5}, {"鸡蛋回锅调味", 3}},
		},
		{
			name: "萝卜炖牛腩", description: "萝卜吸满汤汁，冬天很舒服。", method: "炖", tags: []string{"家常", "滋补", "咸鲜"}, minutes: 95, servings: 4, priceRange: "35-60元", difficulty: 3, proficiency: 3,
			ingredients: []ingredientSeed{{"牛腩", "600", "g"}, {"白萝卜", "1", "根"}, {"八角", "2", "个"}, {"姜片", "5", "片"}},
			steps:       []stepSeed{{"牛腩焯水", 10}, {"香料炒香下牛腩", 8}, {"加水炖煮", 60}, {"放萝卜继续炖软", 15}},
		},
		{
			name: "土豆炖鸡块", description: "软糯入味，汤汁拌饭。", method: "炖", tags: []string{"家常", "下饭", "咸鲜"}, minutes: 45, servings: 3, priceRange: "20-35元", difficulty: 2, proficiency: 4,
			ingredients: []ingredientSeed{{"鸡腿", "2", "只"}, {"土豆", "2", "个"}, {"胡萝卜", "1", "根"}, {"姜片", "4", "片"}},
			steps:       []stepSeed{{"鸡块焯水", 8}, {"煎炒鸡块上色", 6}, {"加水炖煮", 20}, {"加土豆胡萝卜收汁", 10}},
		},
		{
			name: "水煮肉片", description: "麻辣过瘾，适合多人分着吃。", method: "煮", tags: []string{"麻辣", "川味", "聚餐"}, minutes: 35, servings: 4, priceRange: "25-45元", difficulty: 4, proficiency: 3,
			ingredients: []ingredientSeed{{"里脊肉", "350", "g"}, {"豆芽", "300", "g"}, {"青菜", "200", "g"}, {"豆瓣酱", "1", "勺"}, {"花椒", "适量", ""}},
			steps:       []stepSeed{{"肉片腌制上浆", 10}, {"配菜烫熟垫底", 6}, {"炒香底料加汤", 8}, {"滑入肉片煮熟", 5}, {"撒辣椒花椒热油泼香", 4}},
		},
		{
			name: "京酱肉丝", description: "酱香浓郁，卷豆皮很好吃。", method: "炒", tags: []string{"酱香", "家常", "聚餐"}, minutes: 25, servings: 3, priceRange: "18-30元", difficulty: 3, proficiency: 3,
			ingredients: []ingredientSeed{{"里脊肉", "300", "g"}, {"甜面酱", "2", "勺"}, {"大葱", "1", "根"}, {"豆皮", "2", "张"}},
			steps:       []stepSeed{{"肉丝腌制", 8}, {"炒熟肉丝盛出", 5}, {"炒香甜面酱", 4}, {"肉丝回锅裹酱", 4}, {"配葱丝豆皮装盘", 3}},
		},
		{
			name: "油焖大虾", description: "虾壳红亮，咸甜鲜香。", method: "焖", tags: []string{"咸甜", "鲜香", "宴客"}, minutes: 22, servings: 3, priceRange: "35-60元", difficulty: 3, proficiency: 3,
			ingredients: []ingredientSeed{{"大虾", "500", "g"}, {"姜丝", "适量", ""}, {"葱段", "适量", ""}, {"番茄酱", "1", "勺"}},
			steps:       []stepSeed{{"大虾剪须开背", 6}, {"煎出虾油", 6}, {"加调料焖入味", 8}, {"收汁出锅", 2}},
		},
		{
			name: "香菇滑鸡", description: "鸡肉滑嫩，香菇味很足。", method: "蒸", tags: []string{"鲜香", "家常", "嫩滑"}, minutes: 30, servings: 3, priceRange: "20-35元", difficulty: 2, proficiency: 4,
			ingredients: []ingredientSeed{{"鸡腿肉", "350", "g"}, {"干香菇", "8", "朵"}, {"姜丝", "适量", ""}, {"蚝油", "1", "勺"}},
			steps:       []stepSeed{{"香菇泡发切片", 8}, {"鸡肉腌制", 10}, {"铺盘上锅蒸熟", 12}},
		},
		{
			name: "粉蒸肉", description: "软糯不腻，米香很足。", method: "蒸", tags: []string{"家常", "软糯", "宴客"}, minutes: 75, servings: 4, priceRange: "25-45元", difficulty: 3, proficiency: 3,
			ingredients: []ingredientSeed{{"五花肉", "500", "g"}, {"蒸肉米粉", "1", "包"}, {"土豆", "1", "个"}, {"豆瓣酱", "1", "勺"}},
			steps:       []stepSeed{{"五花肉切片腌制", 15}, {"裹蒸肉米粉", 8}, {"土豆垫底铺肉", 5}, {"上锅蒸到软糯", 45}},
		},
		{
			name: "梅菜扣肉", description: "浓香软烂，适合提前准备。", method: "蒸", tags: []string{"咸香", "宴客", "下饭"}, minutes: 120, servings: 5, priceRange: "35-60元", difficulty: 5, proficiency: 2,
			ingredients: []ingredientSeed{{"五花肉", "600", "g"}, {"梅干菜", "150", "g"}, {"老抽", "1", "勺"}, {"冰糖", "少许", ""}},
			steps:       []stepSeed{{"五花肉煮熟抹老抽", 25}, {"肉皮煎上色", 8}, {"切片码碗", 8}, {"炒香梅干菜铺上", 10}, {"上锅蒸到软烂", 65}},
		},
		{
			name: "回锅肉", description: "肥瘦相间，豆瓣酱香。", method: "炒", tags: []string{"微辣", "川味", "下饭"}, minutes: 30, servings: 3, priceRange: "20-35元", difficulty: 3, proficiency: 4,
			ingredients: []ingredientSeed{{"五花肉", "350", "g"}, {"青蒜", "1", "把"}, {"豆瓣酱", "1", "勺"}, {"甜面酱", "半", "勺"}},
			steps:       []stepSeed{{"五花肉煮到断生", 18}, {"切薄片煸出油", 5}, {"炒香豆瓣酱", 3}, {"下青蒜翻匀", 4}},
		},
		{
			name: "蚝油生菜", description: "清爽脆嫩，配饭配粥都行。", method: "炒", tags: []string{"清淡", "素菜", "快手"}, minutes: 8, servings: 2, priceRange: "10元以内", difficulty: 1, proficiency: 5,
			ingredients: []ingredientSeed{{"生菜", "2", "颗"}, {"蒜末", "适量", ""}, {"蚝油", "1", "勺"}, {"生抽", "1", "勺"}},
			steps:       []stepSeed{{"生菜洗净焯水", 3}, {"蒜末炒香调汁", 3}, {"淋汁装盘", 2}},
		},
		{
			name: "手撕包菜", description: "锅气足，酸辣开胃。", method: "炒", tags: []string{"酸辣", "素菜", "快手"}, minutes: 12, servings: 2, priceRange: "10元以内", difficulty: 1, proficiency: 4,
			ingredients: []ingredientSeed{{"包菜", "半", "颗"}, {"干辣椒", "4", "个"}, {"蒜片", "适量", ""}, {"陈醋", "1", "勺"}},
			steps:       []stepSeed{{"包菜撕块控水", 3}, {"爆香蒜片辣椒", 2}, {"大火快炒包菜", 5}, {"沿锅边淋醋", 2}},
		},
		{
			name: "香煎豆腐", description: "外焦里嫩，蘸汁很香。", method: "煎", tags: []string{"家常", "素菜", "香煎"}, minutes: 18, servings: 2, priceRange: "10元以内", difficulty: 2, proficiency: 4,
			ingredients: []ingredientSeed{{"老豆腐", "1", "块"}, {"鸡蛋", "1", "个"}, {"淀粉", "适量", ""}, {"蒜末", "适量", ""}},
			steps:       []stepSeed{{"豆腐切片裹蛋液淀粉", 6}, {"小火煎到两面金黄", 8}, {"调蘸汁淋上", 4}},
		},
		{
			name: "红烧茄子", description: "软糯入味，咸鲜微甜。", method: "烧", tags: []string{"家常", "素菜", "下饭"}, minutes: 25, servings: 2, priceRange: "10-18元", difficulty: 2, proficiency: 4,
			ingredients: []ingredientSeed{{"茄子", "2", "根"}, {"蒜末", "适量", ""}, {"青椒", "1", "个"}, {"生抽", "2", "勺"}},
			steps:       []stepSeed{{"茄子切条煎软", 10}, {"炒香蒜末青椒", 4}, {"倒入酱汁焖入味", 8}, {"收汁出锅", 3}},
		},
		{
			name: "番茄鸡蛋汤", description: "酸甜清爽，十分钟上桌。", method: "汤", tags: []string{"清淡", "快手", "家常"}, minutes: 10, servings: 2, priceRange: "10元以内", difficulty: 1, proficiency: 5,
			ingredients: []ingredientSeed{{"番茄", "1", "个"}, {"鸡蛋", "2", "个"}, {"葱花", "少许", ""}, {"盐", "少许", ""}},
			steps:       []stepSeed{{"番茄炒出汁", 4}, {"加水煮开", 3}, {"淋入蛋液调味", 3}},
		},
		{
			name: "紫菜蛋花汤", description: "清淡鲜美，早餐晚餐都方便。", method: "汤", tags: []string{"清淡", "快手", "鲜美"}, minutes: 8, servings: 2, priceRange: "10元以内", difficulty: 1, proficiency: 5,
			ingredients: []ingredientSeed{{"紫菜", "1", "小把"}, {"鸡蛋", "1", "个"}, {"虾皮", "少许", ""}, {"葱花", "少许", ""}},
			steps:       []stepSeed{{"水开下紫菜虾皮", 3}, {"淋入蛋液", 2}, {"调味撒葱花", 2}},
		},
		{
			name: "莲藕排骨汤", description: "汤清味甜，适合周末慢炖。", method: "汤", tags: []string{"滋补", "清淡", "家常"}, minutes: 100, servings: 4, priceRange: "35-60元", difficulty: 2, proficiency: 3,
			ingredients: []ingredientSeed{{"排骨", "600", "g"}, {"莲藕", "1", "节"}, {"姜片", "4", "片"}, {"枸杞", "少许", ""}},
			steps:       []stepSeed{{"排骨焯水洗净", 10}, {"莲藕切块", 5}, {"加水小火炖煮", 75}, {"加盐和枸杞", 5}},
		},
		{
			name: "山药排骨汤", description: "温和清润，老人孩子都能吃。", method: "汤", tags: []string{"清淡", "滋补", "家常"}, minutes: 85, servings: 4, priceRange: "35-55元", difficulty: 2, proficiency: 3,
			ingredients: []ingredientSeed{{"排骨", "500", "g"}, {"山药", "1", "根"}, {"玉米", "1", "根"}, {"姜片", "4", "片"}},
			steps:       []stepSeed{{"排骨焯水", 10}, {"玉米切段同炖", 50}, {"加入山药炖软", 20}, {"调味出锅", 3}},
		},
		{
			name: "咖喱鸡饭", description: "浓稠咖喱汁，盖饭很方便。", method: "炖", tags: []string{"咖喱", "下饭", "家常"}, minutes: 35, servings: 3, priceRange: "20-35元", difficulty: 2, proficiency: 4,
			ingredients: []ingredientSeed{{"鸡腿肉", "350", "g"}, {"土豆", "1", "个"}, {"胡萝卜", "1", "根"}, {"咖喱块", "2", "块"}},
			steps:       []stepSeed{{"鸡肉和蔬菜切块", 8}, {"鸡肉煎香", 5}, {"加水炖蔬菜", 12}, {"放咖喱块煮浓稠", 8}},
		},
		{
			name: "照烧鸡腿", description: "甜咸酱汁浓，适合便当。", method: "煎", tags: []string{"甜咸", "便当", "家常"}, minutes: 25, servings: 2, priceRange: "18-30元", difficulty: 2, proficiency: 4,
			ingredients: []ingredientSeed{{"鸡腿", "2", "只"}, {"生抽", "2", "勺"}, {"蜂蜜", "1", "勺"}, {"料酒", "1", "勺"}},
			steps:       []stepSeed{{"鸡腿去骨腌制", 10}, {"鸡皮朝下煎出油", 8}, {"倒入照烧汁收浓", 6}},
		},
		{
			name: "葱爆羊肉", description: "葱香浓，羊肉嫩。", method: "爆", tags: []string{"葱香", "快手", "下饭"}, minutes: 15, servings: 2, priceRange: "30-45元", difficulty: 3, proficiency: 3,
			ingredients: []ingredientSeed{{"羊肉片", "300", "g"}, {"大葱", "2", "根"}, {"孜然粉", "少许", ""}, {"生抽", "1", "勺"}},
			steps:       []stepSeed{{"羊肉片快速滑炒", 5}, {"下大葱爆香", 4}, {"调味翻匀出锅", 3}},
		},
		{
			name: "孜然牛肉", description: "香辣有嚼劲，像小馆子味道。", method: "炒", tags: []string{"香辣", "孜然", "下饭"}, minutes: 20, servings: 2, priceRange: "30-45元", difficulty: 3, proficiency: 3,
			ingredients: []ingredientSeed{{"牛肉", "300", "g"}, {"洋葱", "半", "个"}, {"孜然粒", "1", "勺"}, {"辣椒粉", "少许", ""}},
			steps:       []stepSeed{{"牛肉切片腌制", 8}, {"洋葱炒香", 3}, {"牛肉大火快炒", 5}, {"加孜然辣椒翻匀", 3}},
		},
		{
			name: "椒盐排条", description: "外酥里嫩，适合当下酒菜。", method: "炸", tags: []string{"香脆", "椒盐", "聚餐"}, minutes: 35, servings: 3, priceRange: "25-40元", difficulty: 4, proficiency: 2,
			ingredients: []ingredientSeed{{"猪里脊", "350", "g"}, {"鸡蛋", "1", "个"}, {"淀粉", "80", "g"}, {"椒盐", "适量", ""}},
			steps:       []stepSeed{{"里脊切条腌制", 10}, {"挂糊炸定型", 10}, {"复炸到酥脆", 5}, {"撒椒盐装盘", 2}},
		},
		{
			name: "干锅花菜", description: "花菜脆香，微辣下饭。", method: "炒", tags: []string{"微辣", "素菜", "下饭"}, minutes: 20, servings: 2, priceRange: "12-22元", difficulty: 2, proficiency: 4,
			ingredients: []ingredientSeed{{"花菜", "1", "颗"}, {"五花肉片", "80", "g"}, {"干辣椒", "4", "个"}, {"蒜片", "适量", ""}},
			steps:       []stepSeed{{"花菜焯水断生", 4}, {"五花肉煸出油", 5}, {"下花菜大火炒香", 8}, {"调味出锅", 3}},
		},
		{
			name: "虎皮青椒", description: "青椒软香，酱汁下饭。", method: "煎", tags: []string{"家常", "微辣", "素菜"}, minutes: 15, servings: 2, priceRange: "10元以内", difficulty: 2, proficiency: 4,
			ingredients: []ingredientSeed{{"青椒", "6", "个"}, {"蒜末", "适量", ""}, {"生抽", "1", "勺"}, {"香醋", "1", "勺"}},
			steps:       []stepSeed{{"青椒拍裂去籽", 3}, {"干煎出虎皮纹", 6}, {"加入酱汁焖入味", 5}},
		},
		{
			name: "菠菜拌粉丝", description: "清爽小凉菜，酸香开胃。", method: "拌", tags: []string{"清淡", "凉菜", "酸香"}, minutes: 15, servings: 2, priceRange: "10元以内", difficulty: 1, proficiency: 4,
			ingredients: []ingredientSeed{{"菠菜", "250", "g"}, {"粉丝", "1", "把"}, {"蒜末", "适量", ""}, {"香醋", "1", "勺"}},
			steps:       []stepSeed{{"粉丝泡软焯水", 5}, {"菠菜焯水挤干", 4}, {"加入料汁拌匀", 4}},
		},
		{
			name: "酸辣土豆丝", description: "脆爽酸辣，家常快手菜。", method: "炒", tags: []string{"酸辣", "快手", "素菜"}, minutes: 12, servings: 2, priceRange: "10元以内", difficulty: 2, proficiency: 5,
			ingredients: []ingredientSeed{{"土豆", "2", "个"}, {"干辣椒", "3", "个"}, {"白醋", "1", "勺"}, {"蒜片", "适量", ""}},
			steps:       []stepSeed{{"土豆切丝清水洗淀粉", 5}, {"爆香蒜片辣椒", 2}, {"大火快炒土豆丝", 4}, {"淋醋调味", 1}},
		},
		{
			name: "韭菜炒香干", description: "香干有嚼劲，韭菜提香。", method: "炒", tags: []string{"家常", "素菜", "快手"}, minutes: 12, servings: 2, priceRange: "10-15元", difficulty: 1, proficiency: 5,
			ingredients: []ingredientSeed{{"韭菜", "1", "把"}, {"香干", "4", "片"}, {"生抽", "1", "勺"}, {"小米椒", "1", "个"}},
			steps:       []stepSeed{{"韭菜切段香干切条", 4}, {"香干煎香", 4}, {"下韭菜快炒调味", 3}},
		},
		{
			name: "小炒黄牛肉", description: "香辣下饭，牛肉要快炒。", method: "炒", tags: []string{"香辣", "湘味", "下饭"}, minutes: 20, servings: 2, priceRange: "35-55元", difficulty: 4, proficiency: 3,
			ingredients: []ingredientSeed{{"黄牛肉", "300", "g"}, {"香菜", "1", "把"}, {"小米椒", "4", "个"}, {"泡椒", "适量", ""}},
			steps:       []stepSeed{{"牛肉切薄片腌制", 8}, {"辣椒蒜末炒香", 4}, {"牛肉大火快炒", 4}, {"下香菜翻匀", 2}},
		},
		{
			name: "红烧肉", description: "肥而不腻，入口软糯。", method: "烧", tags: []string{"红烧", "家常", "宴客"}, minutes: 75, servings: 4, priceRange: "35-60元", difficulty: 4, proficiency: 3,
			ingredients: []ingredientSeed{{"五花肉", "600", "g"}, {"冰糖", "25", "g"}, {"八角", "2", "个"}, {"姜片", "5", "片"}},
			steps:       []stepSeed{{"五花肉切块焯水", 10}, {"煸出油脂", 10}, {"炒糖色裹肉", 5}, {"加水小火焖烧", 45}, {"收汁出锅", 5}},
		},
		{
			name: "剁椒鱼头", description: "鲜辣开胃，聚餐很有气氛。", method: "蒸", tags: []string{"香辣", "湘味", "宴客"}, minutes: 35, servings: 4, priceRange: "35-60元", difficulty: 4, proficiency: 2,
			ingredients: []ingredientSeed{{"鱼头", "1", "个"}, {"剁椒", "3", "勺"}, {"姜蒜", "适量", ""}, {"葱花", "适量", ""}},
			steps:       []stepSeed{{"鱼头洗净腌制", 10}, {"铺剁椒姜蒜", 5}, {"上锅蒸熟", 15}, {"热油泼香葱花", 3}},
		},
		{
			name: "虾仁滑蛋", description: "嫩滑鲜香，早餐晚餐都合适。", method: "炒", tags: []string{"鲜香", "快手", "嫩滑"}, minutes: 12, servings: 2, priceRange: "20-35元", difficulty: 2, proficiency: 4,
			ingredients: []ingredientSeed{{"虾仁", "200", "g"}, {"鸡蛋", "4", "个"}, {"牛奶", "1", "勺"}, {"葱花", "少许", ""}},
			steps:       []stepSeed{{"虾仁腌制后煎熟", 4}, {"蛋液加牛奶打散", 2}, {"小火滑炒到凝固", 5}},
		},
		{
			name: "黑椒牛柳", description: "黑椒香浓，洋葱脆甜。", method: "炒", tags: []string{"黑椒", "下饭", "家常"}, minutes: 22, servings: 2, priceRange: "30-45元", difficulty: 3, proficiency: 3,
			ingredients: []ingredientSeed{{"牛里脊", "300", "g"}, {"洋葱", "半", "个"}, {"青椒", "1", "个"}, {"黑椒汁", "2", "勺"}},
			steps:       []stepSeed{{"牛肉切条腌制", 8}, {"配菜炒断生", 4}, {"牛柳快炒", 5}, {"加黑椒汁翻匀", 3}},
		},
		{
			name: "南瓜蒸排骨", description: "南瓜软甜，排骨鲜香。", method: "蒸", tags: []string{"家常", "软糯", "咸鲜"}, minutes: 45, servings: 3, priceRange: "25-45元", difficulty: 3, proficiency: 3,
			ingredients: []ingredientSeed{{"排骨", "450", "g"}, {"南瓜", "300", "g"}, {"豆豉", "1", "勺"}, {"蒜末", "适量", ""}},
			steps:       []stepSeed{{"排骨加豆豉蒜末腌制", 15}, {"南瓜切块垫底", 5}, {"排骨铺上蒸熟", 25}},
		},
		{
			name: "卤鸡腿", description: "一次多卤几个，冷吃也好吃。", method: "卤", tags: []string{"卤香", "便当", "家常"}, minutes: 60, servings: 4, priceRange: "25-40元", difficulty: 2, proficiency: 4,
			ingredients: []ingredientSeed{{"鸡腿", "6", "只"}, {"卤料包", "1", "包"}, {"生抽", "3", "勺"}, {"冰糖", "少许", ""}},
			steps:       []stepSeed{{"鸡腿焯水", 8}, {"卤汁煮开", 5}, {"小火卤入味", 35}, {"关火浸泡", 10}},
		},
		{
			name: "酱牛肉", description: "切片下酒，做一次能吃几顿。", method: "卤", tags: []string{"酱香", "凉菜", "宴客"}, minutes: 120, servings: 5, priceRange: "60-90元", difficulty: 4, proficiency: 2,
			ingredients: []ingredientSeed{{"牛腱子", "800", "g"}, {"黄豆酱", "2", "勺"}, {"八角", "2", "个"}, {"桂皮", "1", "段"}},
			steps:       []stepSeed{{"牛腱浸泡去血水", 20}, {"焯水洗净", 10}, {"加酱料小火卤煮", 70}, {"浸泡放凉切片", 20}},
		},
		{
			name: "香煎带鱼", description: "外皮焦香，鱼肉细嫩。", method: "煎", tags: []string{"香煎", "家常", "咸鲜"}, minutes: 25, servings: 3, priceRange: "25-40元", difficulty: 3, proficiency: 3,
			ingredients: []ingredientSeed{{"带鱼段", "500", "g"}, {"姜丝", "适量", ""}, {"料酒", "1", "勺"}, {"淀粉", "适量", ""}},
			steps:       []stepSeed{{"带鱼腌制去腥", 10}, {"表面拍薄粉", 3}, {"小火煎到两面金黄", 10}},
		},
		{
			name: "糖醋藕丁", description: "酸甜脆爽，素菜也很下饭。", method: "炒", tags: []string{"酸甜", "素菜", "快手"}, minutes: 15, servings: 2, priceRange: "10-18元", difficulty: 2, proficiency: 4,
			ingredients: []ingredientSeed{{"莲藕", "1", "节"}, {"青椒", "半", "个"}, {"白醋", "1", "勺"}, {"白糖", "1", "勺"}},
			steps:       []stepSeed{{"莲藕切丁焯水", 5}, {"调糖醋汁", 2}, {"藕丁快炒", 5}, {"倒汁收亮", 3}},
		},
		{
			name: "香辣鸡爪", description: "软糯入味，追剧小菜。", method: "焖", tags: []string{"香辣", "下酒", "小吃"}, minutes: 55, servings: 3, priceRange: "25-40元", difficulty: 3, proficiency: 3,
			ingredients: []ingredientSeed{{"鸡爪", "600", "g"}, {"小米椒", "4", "个"}, {"蒜", "半", "头"}, {"生抽", "3", "勺"}},
			steps:       []stepSeed{{"鸡爪剪甲焯水", 10}, {"炒香蒜末辣椒", 4}, {"下鸡爪加水焖煮", 35}, {"收汁入味", 5}},
		},
		{
			name: "上汤娃娃菜", description: "汤鲜菜甜，清淡不寡淡。", method: "煮", tags: []string{"清淡", "鲜美", "素菜"}, minutes: 18, servings: 2, priceRange: "10-20元", difficulty: 2, proficiency: 4,
			ingredients: []ingredientSeed{{"娃娃菜", "2", "颗"}, {"皮蛋", "1", "个"}, {"火腿", "少许", ""}, {"蒜末", "适量", ""}},
			steps:       []stepSeed{{"娃娃菜切条", 3}, {"蒜末皮蛋火腿炒香", 4}, {"加水煮成上汤", 5}, {"下娃娃菜煮软", 5}},
		},
		{
			name: "豉汁蒸排骨", description: "排骨嫩滑，豆豉香浓。", method: "蒸", tags: []string{"咸鲜", "豆豉", "家常"}, minutes: 40, servings: 3, priceRange: "25-45元", difficulty: 3, proficiency: 3,
			ingredients: []ingredientSeed{{"肋排", "500", "g"}, {"豆豉", "1", "勺"}, {"蒜末", "适量", ""}, {"淀粉", "1", "勺"}},
			steps:       []stepSeed{{"排骨泡水去血水", 10}, {"加豆豉蒜末腌制", 15}, {"上锅蒸熟", 15}},
		},
	}
}
