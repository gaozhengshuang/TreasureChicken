-- Generated by github.com/davyxu/tabtoy
-- Version: 2.8.10

local tab = {
	TSign = {
		{ Id = 1, CostId = 6005, Num = 10 	},
		{ Id = 2, CostId = 6005, Num = 10 	},
		{ Id = 3, CostId = 6005, Num = 10 	},
		{ Id = 4, CostId = 6005, Num = 10 	},
		{ Id = 5, CostId = 4027, Num = 1 	},
		{ Id = 6, CostId = 6005, Num = 10 	},
		{ Id = 7, CostId = 6005, Num = 10 	}
	}

}


-- Id
tab.TSignById = {}
for _, rec in pairs(tab.TSign) do
	tab.TSignById[rec.Id] = rec
end

tab.Enum = {
}

return tab